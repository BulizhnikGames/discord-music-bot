package bot

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"github.com/jogramming/dca"
	"io"
	"log"
	"time"
)

const PLAYBACK_TIMEOUT = 30 * time.Minute

func (voiceChat *VoiceEntity) PlaySongs(ctx context.Context, bot *DiscordBot) {
	defer func() {
		_ = bot.LeaveVoiceChat(voiceChat.voiceConnection.GuildID)
	}()

	timeout := time.NewTimer(PLAYBACK_TIMEOUT).C
	for {
		select {
		case <-timeout:
			log.Printf("Playback timeout (channel: %s, guild: %s)", voiceChat.voiceConnection.GuildID, voiceChat.voiceConnection.GuildID)
			return
		case <-ctx.Done():
			log.Printf("leave signal (channel: %s, guild: %s)", voiceChat.voiceConnection.GuildID, voiceChat.voiceConnection.GuildID)
			return
		case <-voiceChat.queue.NewHandled:
			song := voiceChat.queue.ReadHandled()
			if song == nil {
				continue
			}
			if song.FilePath == "" {
				log.Printf("couldn't play song")
				voiceChat.SendErrorPlaybackMessage(bot.Session, *song)
				continue
			}
			voiceChat.mutex.Lock()
			if voiceChat.playbackContext == nil || voiceChat.playbackContext.Err() != nil {
				playbackCtx, cancel := context.WithCancel(ctx)
				voiceChat.playbackContext = playbackCtx
				voiceChat.forceStop = func(setText string) {
					voiceChat.mutex.Lock()
					defer voiceChat.mutex.Unlock()

					if setText == "" {
						voiceChat.DeletePlaybackMessage(bot.Session)
					} else {
						voiceChat.mutex.Unlock()
						_ = voiceChat.SetPlaybackMessageToText(bot.Session, setText)
						voiceChat.mutex.Lock()
					}

					cancel()

					if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.EncodeSession != nil {
						voiceChat.nowPlaying.EncodeSession.Cleanup()
						voiceChat.nowPlaying = nil
					}
				}
			}
			voiceChat.mutex.Unlock()
			log.Printf("Playing song %s (query %s)", song.Title, song.Query)
			err := voiceChat.playSong(voiceChat.playbackContext, bot.Session, song)
			if err != nil {
				log.Printf("Error playing song: %s", err.Error())
			}
			timeout = time.NewTimer(PLAYBACK_TIMEOUT).C
		}
	}
}

func (voiceChat *VoiceEntity) playSong(ctx context.Context, session *discordgo.Session, song *internal.Song) error {
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
	options.Path = config.Utils
	options.Bitrate = 96
	options.RawOutput = true

	encodeSession, err := dca.EncodeFile(song.FilePath, options)
	if err != nil {
		voiceChat.SendErrorPlaybackMessage(session, *song)
		return errors.Errorf("Failed to create encoding session for %s: %v", song.FilePath, err)
	}
	defer encodeSession.Cleanup()

	playContext, cancel := context.WithCancel(ctx)
	voiceChat.mutex.Lock()
	voiceChat.nowPlaying = &internal.PlayingSong{
		Song:          song,
		EncodeSession: encodeSession,
	}
	voiceChat.nowPlaying.Skip = func(setText string) {
		voiceChat.mutex.Lock()
		defer voiceChat.mutex.Unlock()

		// delete only if queue is empty otherwise notify user about it
		if setText == "" && voiceChat.queue.Len > 0 {
			voiceChat.DeletePlaybackMessage(session)
		} else {
			if setText == "" {
				setText = "🔇  queue has ended!  🔇"
			}
			voiceChat.mutex.Unlock()
			_ = voiceChat.SetPlaybackMessageToText(session, setText)
			voiceChat.mutex.Lock()
		}

		voiceChat.nowPlaying = nil
		cancel()
		if voiceChat.loop == 2 {
			return
		}

		voiceChat.cache.Mutex.Lock()
		defer voiceChat.cache.Mutex.Unlock()
		cache, ok := voiceChat.cache.Data[song.Query]
		if ok {
			cache.Cnt--
			if voiceChat.loop == 0 && cache.Cnt <= 0 {
				encodeSession.Cleanup()
				cache.Delete()
				delete(voiceChat.cache.Data, song.Query)
			}
		}

		if voiceChat.loop == 1 {
			voiceChat.queue.Write(*song)
		}
	}

	done := make(chan error)
	voiceChat.nowPlaying.Stream = dca.NewStream(encodeSession, voiceChat.voiceConnection, done)
	voiceChat.mutex.Unlock()

	err = voiceChat.NewPlaybackMessage(session)
	if err != nil {
		return errors.Errorf("couldn't send playback message: %v", err)
	}

	select {
	case err = <-done:
		if err != nil && err != io.EOF {
			log.Printf("error while streaming (song %s): %v", song.Title, err)
		}
	case <-playContext.Done():
		log.Printf("Skipped %s", song.Title)
		voiceChat.mutex.RLock()
		if voiceChat.loop == 2 {
			voiceChat.mutex.RUnlock()
			encodeSession.Cleanup()
			return voiceChat.playSong(ctx, session, song)
		}
		voiceChat.mutex.RUnlock()
		return nil
	}
	voiceChat.nowPlaying.Skip("")
	log.Printf("End of song %s", song.Title)
	voiceChat.mutex.RLock()
	if voiceChat.loop == 2 {
		voiceChat.mutex.RUnlock()
		encodeSession.Cleanup()
		return voiceChat.playSong(ctx, session, song)
	}
	voiceChat.mutex.RUnlock()
	return nil
}

func (voiceChat *VoiceEntity) NewPlaybackMessage(session *discordgo.Session) error {
	voiceChat.mutex.Lock()
	defer voiceChat.mutex.Unlock()

	if voiceChat.nowPlaying == nil || voiceChat.nowPlaying.Song == nil {
		return errors.New("no song playing")
	}
	song := voiceChat.nowPlaying.Song

	firstLine, err := discordgo.MessageComponentFromJSON(voiceChat.constructJsonLine(
		voiceChat.pauseButtonJson(),
		voiceChat.skipButtonJson(),
		voiceChat.stopButtonJson(),
	))
	if err != nil {
		return err
	}

	secondLine, err := discordgo.MessageComponentFromJSON(voiceChat.constructJsonLine(
		voiceChat.shuffleQueueJson(),
		voiceChat.loopOptsJson(),
	))
	if err != nil {
		return err
	}

	var title, author string
	if song.Title == "" {
		title = "Loading..."
	} else {
		title = fmt.Sprintf("%s - [%d:%02d]", song.Title, song.Duration/60, song.Duration%60)
	}
	if song.Author == "" {
		author = "Loading..."
	} else {
		author = song.Author
	}

	msg, err := session.ChannelMessageSendComplex(voiceChat.textChannel,
		&discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    "Now Playing",
						IconURL: session.State.User.AvatarURL("64x64"),
					},
					Title: title,
					URL:   song.URL,
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
	voiceChat.nowPlaying.Message = msg
	return nil
}

func (voiceChat *VoiceEntity) TryRegenPlaybackMessage(session *discordgo.Session) {
	voiceChat.mutex.Lock()
	defer voiceChat.mutex.Unlock()

	var id, channel = "", ""
	if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.Message != nil {
		id = voiceChat.nowPlaying.Message.ID
		channel = voiceChat.nowPlaying.Message.ChannelID
	}

	var s = internal.Song{}
	if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.Song != nil {
		s = *voiceChat.nowPlaying.Song
	}

	if id == "" {
		go voiceChat.NewPlaybackMessage(session)
	} else {
		//log.Printf(id)
		firstLine, err := discordgo.MessageComponentFromJSON(voiceChat.constructJsonLine(
			voiceChat.pauseButtonJson(),
			voiceChat.skipButtonJson(),
			voiceChat.stopButtonJson(),
		))
		if err != nil {
			return
		}
		secondLine, err := discordgo.MessageComponentFromJSON(voiceChat.constructJsonLine(
			voiceChat.shuffleQueueJson(),
			voiceChat.loopOptsJson(),
		))
		if err != nil {
			return
		}

		var title, author string
		if s.Title == "" {
			title = "Loading..."
		} else {
			title = fmt.Sprintf("%s - [%d:%02d]", s.Title, s.Duration/60, s.Duration%60)
		}
		if s.Author == "" {
			author = "Loading..."
		} else {
			author = s.Author
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
					URL:   s.URL,
					Color: 2326507,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: author,
						},
					},
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: s.ThumbnailUrl,
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
		voiceChat.nowPlaying.Message = msg
	}
}

func (voiceChat *VoiceEntity) SetPlaybackMessageToText(session *discordgo.Session, text string) error {
	voiceChat.mutex.Lock()
	defer voiceChat.mutex.Unlock()

	var id, channel = "", ""
	if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.Message != nil {
		id = voiceChat.nowPlaying.Message.ID
		channel = voiceChat.nowPlaying.Message.ChannelID
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
	voiceChat.nowPlaying.Message = msg
	return nil
}

func (voiceChat *VoiceEntity) DeletePlaybackMessage(session *discordgo.Session) {
	if voiceChat.nowPlaying.Message != nil && len(voiceChat.nowPlaying.Message.Components) > 0 {
		_ = session.ChannelMessageDelete(voiceChat.nowPlaying.Message.ChannelID, voiceChat.nowPlaying.Message.ID)
	}
}

func (voiceChat *VoiceEntity) SendErrorPlaybackMessage(session *discordgo.Session, song internal.Song) {
	var message string
	if song.Title != "" && song.Author != "" {
		message = fmt.Sprintf(
			"❌  couldn't play song %s by %s  ❌",
			song.Title,
			song.Author,
		)
	} else {
		message = fmt.Sprintf(
			"❌  couldn't play song %s  ❌",
			song.Query,
		)
	}
	_, err := session.ChannelMessageSendComplex(voiceChat.textChannel,
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
		log.Printf("Couldn't send error playback message: %v", err)
	}
}
