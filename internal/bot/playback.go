package bot

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/go-faster/errors"
	"github.com/jogramming/dca"
	"io"
	"log"
	"time"
)

const PLAYBACK_TIMEOUT = 30 * time.Minute

func (voiceChat *VoiceEntity) PlaySongs(ctx context.Context, bot *DiscordBot) {
	defer func() {
		err := bot.LeaveVoiceChat(voiceChat.voiceConnection.GuildID)
		if err != nil {
			log.Printf(
				"Couldn't leave voice chat (id: %s, guild: %s): %v",
				voiceChat.voiceConnection.ChannelID,
				voiceChat.voiceConnection.GuildID,
				err,
			)
		}
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
		default:
			song := voiceChat.Queue.ReadHandled()
			if song == nil {
				continue
			}
			message := fmt.Sprintf(
				":arrow_forward: playing song: `%s | %d:%02d` by `%s`",
				song.Title,
				song.Duration/60,
				song.Duration%60,
				song.Author,
			)
			err := bot.SendInChannel(voiceChat.textChannel, message)
			if err != nil {
				log.Printf("Couldn't send message about song: %v", err)
				continue
			}
			log.Printf("Playing song %s", song.Title)
			err = voiceChat.playSong(ctx, song)
			if err != nil {
				log.Printf("Error playing song: %v", err)
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
	voiceChat.skip = func() {
		voiceChat.nowPlaying = nil
		cancel()
		if voiceChat.loop == 0 {
			voiceChat.cache.Mutex.Lock()
			defer voiceChat.cache.Mutex.Unlock()
			if cache, ok := voiceChat.cache.Data[song.Query]; ok {
				cache.Cnt--
				if cache.Cnt == 0 {
					encodeSession.Cleanup()
					cache.Delete()
					delete(voiceChat.cache.Data, song.Query)
				}
			}
		} else if voiceChat.loop == 1 {
			voiceChat.InsertQueue(song.Query)
		}
	}

	done := make(chan error)
	voiceChat.nowPlaying = song
	dca.NewStream(encodeSession, voiceChat.voiceConnection, done)
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
	voiceChat.skip()
	log.Printf("End of song %s", song.Title)
	if voiceChat.loop == 2 {
		encodeSession.Cleanup()
		return voiceChat.playSong(ctx, song)
	}
	return nil
}
