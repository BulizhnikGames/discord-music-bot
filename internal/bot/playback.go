package bot

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
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
			log.Printf("stop signal (channel: %s, guild: %s)", voiceChat.voiceConnection.GuildID, voiceChat.voiceConnection.GuildID)
			return
		case <-voiceChat.queue.NewHandled:
			song := voiceChat.queue.ReadHandled()
			if song == nil {
				continue
			}
			var message string
			if song.FilePath == "" {
				if song.Title != "" && song.Duration != 0 && song.Author != "" {
					message = fmt.Sprintf(
						":x: couldn't play song `%s | %d:%02d` by `%s` :x: ",
						song.Title,
						song.Duration/60,
						song.Duration%60,
						song.Author,
					)
				} else {
					message = fmt.Sprintf(
						":x: couldn't play song `%s` :x: ",
						song.Query,
					)
				}
				err := bot.SendInChannel(voiceChat.textChannel, message)
				if err != nil {
					log.Printf("Couldn't send error message about song: %v", err)
				}
				continue
			}
			var err error
			msg, err := voiceChat.SendPlayBackMessage(bot, *song)
			if err != nil {
				log.Printf("Couldn't send new playback message: %s", err.Error())
				continue
			}
			if voiceChat.playbackMessage != nil {
				err = bot.Session.ChannelMessageDelete(voiceChat.textChannel, voiceChat.playbackMessage.ID)
				if err != nil {
					log.Printf("Couldn't delete old playbak message: %s", err.Error())
				}
			}
			voiceChat.playbackMessage = msg
			log.Printf("Playing song %s (query %s)", song.Title, song.Query)
			err = voiceChat.playSong(ctx, song)
			if err != nil {
				log.Printf("Error playing song: %s", err.Error())
			}
			timeout = time.NewTimer(PLAYBACK_TIMEOUT).C
		}
	}
}

func (voiceChat *VoiceEntity) playSong(ctx context.Context, song *internal.Song) error {
	options := dca.StdEncodeOptions
	options.BufferedFrames = 100
	options.FrameDuration = 20
	//options.CompressionLevel = 5
	options.Path = config.Utils
	options.Bitrate = 96
	options.RawOutput = true

	//log.Printf("%+v", song)

	encodeSession, err := dca.EncodeFile(song.FilePath, options)
	if err != nil {
		return errors.Errorf("Failed to create encoding session for %s: %v", song.FilePath, err)
	}
	defer encodeSession.Cleanup()

	//time.Sleep(500 * time.Millisecond)

	playContext, cancel := context.WithCancel(ctx)
	voiceChat.nowPlaying = &internal.PlayingSong{
		Song: song,
	}
	voiceChat.nowPlaying.Skip = func(clear bool) {
		if clear {
			encodeSession.Cleanup()
			cancel()
			return
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
			// maybe consider this variant
			//voiceChat.queue.Write(song)
			voiceChat.InsertQueue(*song)
		}
	}

	done := make(chan error)
	voiceChat.nowPlaying.Stream = dca.NewStream(encodeSession, voiceChat.voiceConnection, done)
	select {
	case err = <-done:
		if err != nil && err != io.EOF {
			log.Printf("error while streaming (song %s): %v", song.Title, err)
		}
	case <-playContext.Done():
		log.Printf("Skipped %s", song.Title)
		if voiceChat.loop == 2 {
			encodeSession.Cleanup()
			return voiceChat.playSong(ctx, song)
		}
		return nil
	}
	voiceChat.nowPlaying.Skip(false)
	log.Printf("End of song %s", song.Title)
	if voiceChat.loop == 2 {
		encodeSession.Cleanup()
		return voiceChat.playSong(ctx, song)
	}
	return nil
}
