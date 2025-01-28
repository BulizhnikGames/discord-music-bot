package voice

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/dca"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"io"
	"log"
	"sync"
)

type Connection struct {
	VoiceConnection *discordgo.VoiceConnection
	Queue           *internal.MusicQueue
	NowPlaying      *internal.PlayingSong
	Cache           internal.AsyncMap[string, *internal.SongCache] // Key is user's query for song
	Loop            int                                            // 0 - no Loop, 1 - Loop over Queue, 2 - Loop over song
	// ID of channel where first /play command was sent (all next playback messages will be sent there)
	TextChannel     string
	playbackContext context.Context // Context of all playback cancels when Queue is empty by ForceStop
	// Cancel func for playback context, sets playback message's text ot arg
	// Skip function on NowPlaying won't stop playback when Loop = 2, but this will
	ForceStop func(playbackText string)
	Leave     context.CancelFunc // Cancels main context of voice conn
	Mutex     *sync.RWMutex      // Mutex for fields: NowPlaying, Loop, playbackContext, ForceStop

	Logger *log.Logger
}

func (voiceChat *Connection) PlaySongs(ctx context.Context, session *discordgo.Session, updateTimer chan<- struct{}) {
	defer voiceChat.VoiceConnection.Disconnect()

	for {
		select {
		case <-ctx.Done():
			voiceChat.Logger.Printf("Leave signal (channel: %s, guild: %s)", voiceChat.VoiceConnection.GuildID, voiceChat.VoiceConnection.GuildID)
			return
		case <-voiceChat.Queue.NewHandled:
			song := voiceChat.Queue.ReadHandled()
			if song == nil {
				continue
			}
			if song.FileUrl == "" {
				voiceChat.Logger.Printf("couldn't play song: %+v:", song)
				voiceChat.SendErrorPlaybackMessage(session, *song)
				continue
			}
			voiceChat.Mutex.Lock()
			if voiceChat.playbackContext == nil || voiceChat.playbackContext.Err() != nil {
				playbackCtx, cancel := context.WithCancel(ctx)
				voiceChat.playbackContext = playbackCtx
				voiceChat.ForceStop = func(setText string) {
					voiceChat.Mutex.Lock()
					defer voiceChat.Mutex.Unlock()

					if setText == "" {
						voiceChat.DeletePlaybackMessage(session)
					} else {
						voiceChat.Mutex.Unlock()
						_ = voiceChat.SetPlaybackMessageToText(session, setText)
						voiceChat.Mutex.Lock()
					}

					cancel()

					if voiceChat.NowPlaying != nil && voiceChat.NowPlaying.EncodeSession != nil {
						voiceChat.NowPlaying.EncodeSession.Cleanup()
						voiceChat.NowPlaying = nil
					}
				}
			}
			voiceChat.Mutex.Unlock()
			//log.Printf("Playing song %s (%s)", song.Title, song.FileUrl)
			err := voiceChat.playSong(voiceChat.playbackContext, session, song)
			if err != nil {
				voiceChat.Logger.Printf("Error playing song: %s", err.Error())
			}
			updateTimer <- struct{}{}
		}
	}
}

func (voiceChat *Connection) playSong(ctx context.Context, session *discordgo.Session, song *internal.Song) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		break
	}

	options := dca.StdEncodeOptions
	options.BufferedFrames = 100
	options.FrameDuration = 20
	//options.CompressionLevel = 5
	options.Path = config.Tools
	options.Bitrate = 96
	options.RawOutput = true

	encodeSession, err := dca.EncodeFile(song.FileUrl, options)
	if err != nil {
		voiceChat.SendErrorPlaybackMessage(session, *song)
		return errors.Errorf("Failed to create encoding session for %s (%s): %v", song.Query, song.FileUrl, err)
	}
	defer encodeSession.Cleanup()

	playContext, cancel := context.WithCancel(ctx)
	voiceChat.Mutex.Lock()
	voiceChat.NowPlaying = &internal.PlayingSong{
		Song:          song,
		EncodeSession: encodeSession,
	}
	voiceChat.NowPlaying.Skip = func(setText string) {
		voiceChat.Mutex.Lock()
		defer voiceChat.Mutex.Unlock()

		// delete or set text only if queue isn't empty otherwise notify user about it
		if voiceChat.Queue.Len() > 0 || voiceChat.Loop != 0 {
			if setText == "" {
				voiceChat.DeletePlaybackMessage(session)
			} else {
				voiceChat.Mutex.Unlock()
				_ = voiceChat.SetPlaybackMessageToText(session, setText)
				voiceChat.Mutex.Lock()
			}
		} else {
			voiceChat.Mutex.Unlock()
			_ = voiceChat.SetPlaybackMessageToText(session, "🔇  Queue has ended!  🔇")
			voiceChat.Mutex.Lock()
		}

		voiceChat.NowPlaying = nil
		cancel()

		if voiceChat.Loop == 2 {
			return
		}

		voiceChat.Cache.Mutex.Lock()
		defer voiceChat.Cache.Mutex.Unlock()
		cache, ok := voiceChat.Cache.Data[song.Query]
		if ok {
			cache.Cnt--
			if voiceChat.Loop == 0 && cache.Cnt <= 0 {
				//encodeSession.Cleanup()
				delete(voiceChat.Cache.Data, song.Query)
			}
		}

		if voiceChat.Loop == 1 {
			voiceChat.Queue.Write(*song)
		}
	}

	done := make(chan error)
	voiceChat.NowPlaying.Stream = dca.NewStream(encodeSession, voiceChat.VoiceConnection, done)
	voiceChat.Mutex.Unlock()

	err = voiceChat.NewPlaybackMessage(session)
	if err != nil {
		return errors.Errorf("couldn't send playback message: %v", err)
	}

	voiceChat.Logger.Printf("Playing song %s", song.Title)

	select {
	case err = <-done:
		if err != nil && err != io.EOF {
			voiceChat.Logger.Printf("error while streaming (song %s): %v", song.Title, err)
		}
	case <-playContext.Done():
		voiceChat.Logger.Printf("Skipped %s", song.Title)
		voiceChat.Mutex.RLock()
		if voiceChat.Loop == 2 {
			voiceChat.Mutex.RUnlock()
			encodeSession.Cleanup()
			return voiceChat.playSong(ctx, session, song)
		}
		voiceChat.Mutex.RUnlock()
		return nil
	}
	voiceChat.NowPlaying.Skip("")
	voiceChat.Logger.Printf("End of song %s", song.Title)
	voiceChat.Mutex.RLock()
	if voiceChat.Loop == 2 {
		voiceChat.Mutex.RUnlock()
		encodeSession.Cleanup()
		return voiceChat.playSong(ctx, session, song)
	}
	voiceChat.Mutex.RUnlock()
	return nil
}

func (voiceChat *Connection) NewPlaybackMessage(session *discordgo.Session) error {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()

	if voiceChat.NowPlaying == nil || voiceChat.NowPlaying.Song == nil {
		return errors.New("no song playing")
	}
	song := voiceChat.NowPlaying.Song

	firstLine, err := discordgo.MessageComponentFromJSON(ConstructJsonLine(
		voiceChat.pauseButtonJson(session.State.SessionID),
		voiceChat.skipButtonJson(session.State.SessionID),
		voiceChat.stopButtonJson(session.State.SessionID),
	))
	if err != nil {
		return err
	}

	secondLine, err := discordgo.MessageComponentFromJSON(ConstructJsonLine(
		voiceChat.shuffleQueueJson(session.State.SessionID),
		voiceChat.loopOptsJson(session.State.SessionID),
	))
	if err != nil {
		return err
	}

	var title, author = "Loading...", "Loading..."
	if song.Title != "" {
		title = fmt.Sprintf("%s - [%d:%02d]", song.Title, song.Duration/60, song.Duration%60)
	}
	if song.Author != "" {
		author = song.Author
	}

	msg, err := session.ChannelMessageSendComplex(voiceChat.TextChannel,
		&discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    "Now Playing",
						IconURL: session.State.User.AvatarURL("64x64"),
					},
					Title: title,
					URL:   song.OriginalUrl,
					Color: 2326507,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: author,
						},
					},
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: song.ThumbnailUrl,
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "github.com/BulizhnikGames/discord-music-bot",
					},
				},
			},
			Components: []discordgo.MessageComponent{
				firstLine,
				secondLine,
			},
		},
	)
	if err != nil {
		return err
	}
	voiceChat.Mutex.RUnlock()
	voiceChat.Mutex.Lock()
	voiceChat.NowPlaying.Message = msg
	voiceChat.Mutex.Unlock()
	voiceChat.Mutex.RLock()
	return nil
}

func (voiceChat *Connection) TryRegenPlaybackMessage(session *discordgo.Session) {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()

	var id, channel = "", ""
	if voiceChat.NowPlaying != nil && voiceChat.NowPlaying.Message != nil {
		id = voiceChat.NowPlaying.Message.ID
		channel = voiceChat.NowPlaying.Message.ChannelID
	}

	var song = internal.Song{}
	if voiceChat.NowPlaying != nil && voiceChat.NowPlaying.Song != nil {
		song = *voiceChat.NowPlaying.Song
	}

	if id == "" {
		go voiceChat.NewPlaybackMessage(session)
	} else {
		//log.Printf(id)
		firstLine, err := discordgo.MessageComponentFromJSON(ConstructJsonLine(
			voiceChat.pauseButtonJson(session.State.SessionID),
			voiceChat.skipButtonJson(session.State.SessionID),
			voiceChat.stopButtonJson(session.State.SessionID),
		))
		if err != nil {
			return
		}
		secondLine, err := discordgo.MessageComponentFromJSON(ConstructJsonLine(
			voiceChat.shuffleQueueJson(session.State.SessionID),
			voiceChat.loopOptsJson(session.State.SessionID),
		))
		if err != nil {
			return
		}

		var title, author = "Loading...", "Loading..."
		if song.Title != "" {
			title = fmt.Sprintf("%s - [%d:%02d]", song.Title, song.Duration/60, song.Duration%60)
		}
		if song.Author != "" {
			author = song.Author
		}

		msg, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:      id,
			Channel: channel,
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    "Now Playing",
						IconURL: session.State.User.AvatarURL("64x64"),
					},
					Title: title,
					URL:   song.OriginalUrl,
					Color: 2326507,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: author,
						},
					},
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: song.ThumbnailUrl,
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "github.com/BulizhnikGames/discord-music-bot",
					},
				},
			},
			Components: &[]discordgo.MessageComponent{
				firstLine,
				secondLine,
			},
		})
		if err != nil {
			return
		}
		voiceChat.Mutex.RUnlock()
		voiceChat.Mutex.Lock()
		voiceChat.NowPlaying.Message = msg
		voiceChat.Mutex.Unlock()
		voiceChat.Mutex.RLock()
	}
}

func (voiceChat *Connection) SetPlaybackMessageToText(session *discordgo.Session, text string) error {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()

	var id, channel = "", ""
	if voiceChat.NowPlaying != nil && voiceChat.NowPlaying.Message != nil {
		id = voiceChat.NowPlaying.Message.ID
		channel = voiceChat.NowPlaying.Message.ChannelID
	}

	if id == "" {
		return errors.New("no playback message to edit")
	}

	msg, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      id,
		Channel: channel,
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Author: &discordgo.MessageEmbedAuthor{
					Name:    text,
					IconURL: session.State.User.AvatarURL("64x64"),
				},
				Color: 2326507,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "github.com/BulizhnikGames/discord-music-bot",
				},
			},
		},
		Components: &[]discordgo.MessageComponent{},
	})
	if err != nil {
		return err
	}
	voiceChat.Mutex.RUnlock()
	voiceChat.Mutex.Lock()
	voiceChat.NowPlaying.Message = msg
	voiceChat.Mutex.Unlock()
	voiceChat.Mutex.RLock()
	return nil
}

func (voiceChat *Connection) DeletePlaybackMessage(session *discordgo.Session) {
	if voiceChat.NowPlaying.Message != nil && len(voiceChat.NowPlaying.Message.Components) > 0 {
		_ = session.ChannelMessageDelete(voiceChat.NowPlaying.Message.ChannelID, voiceChat.NowPlaying.Message.ID)
	}
}

func (voiceChat *Connection) SendErrorPlaybackMessage(session *discordgo.Session, song internal.Song) {
	var message string
	if song.Title != "" {
		message = fmt.Sprintf(
			"❌  couldn't play song %s  ❌",
			song.Title,
		)
	} else {
		message = fmt.Sprintf(
			"❌  couldn't play song %s  ❌",
			song.Query,
		)
	}
	_, err := session.ChannelMessageSendComplex(voiceChat.TextChannel,
		&discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    message,
						IconURL: session.State.User.AvatarURL("64x64"),
					},
					Color: 15410030,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "github.com/BulizhnikGames/discord-music-bot",
					},
				},
			},
		},
	)
	if err != nil {
		voiceChat.Logger.Printf("Couldn't send error playback message: %v", err)
	}
}
