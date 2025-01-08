package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"github.com/go-faster/errors"
	"github.com/jogramming/dca"
	"io"
	"log"
	"time"
)

const PLAYBACK_TIMEOUT = 30 * time.Minute

func (bot *DiscordBot) LeaveVoiceChat(guildID string) error {
	bot.VoiceEntities.Mutex.Lock()
	defer bot.VoiceEntities.Mutex.Unlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.Stop()
		close(voiceChat.Queue)
		err := voiceChat.VoiceConnection.Disconnect()
		if err != nil {
			return err
		}
		delete(bot.VoiceEntities.Data, guildID)
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) ClearQueue(guildID string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if voiceChat.Skip != nil {
			voiceChat.Skip()
		}
		for len(voiceChat.Queue) > 0 {
			<-voiceChat.Queue
		}
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (voiceChat *VoiceEntity) PlaySongs(ctx context.Context, bot *DiscordBot) {
	defer func() {
		err := bot.LeaveVoiceChat(voiceChat.VoiceConnection.GuildID)
		if err != nil {
			log.Printf(
				"Couldn't leave voice chat (id: %s, guild: %s): %v",
				voiceChat.VoiceConnection.ChannelID,
				voiceChat.VoiceConnection.GuildID,
				err,
			)
		}
	}()

	timeout := time.NewTimer(PLAYBACK_TIMEOUT).C
	for {
		select {
		case song := <-voiceChat.Queue:
			downloaded, err := youtube.Download(song)
			if err != nil {
				log.Printf("Error downloading song: %v", err)
				return
			}
			log.Printf("Playing song %s", song)
			err = voiceChat.playSong(ctx, downloaded)
			if err != nil {
				log.Printf("Error playing song: %v", err)
			}
			timeout = time.NewTimer(PLAYBACK_TIMEOUT).C
		case <-timeout:
			log.Printf("Playback timeout (channel: %s, guild: %s)", voiceChat.VoiceConnection.GuildID, voiceChat.VoiceConnection.GuildID)
			return
		case <-ctx.Done():
			log.Printf("Stop signal (channel: %s, guild: %s)", voiceChat.VoiceConnection.GuildID, voiceChat.VoiceConnection.GuildID)
			return
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

	encodeSession, err := dca.EncodeFile(song.FilePath, options)
	if err != nil {
		return errors.Errorf("Failed to create encoding session for %s: %v", song.FilePath, err)
	}
	defer encodeSession.Cleanup()

	time.Sleep(500 * time.Millisecond)

	playContext, cancel := context.WithCancel(ctx)
	voiceChat.Skip = cancel

	done := make(chan error)
	dca.NewStream(encodeSession, voiceChat.VoiceConnection, done)
	select {
	case err = <-done:
		if err != nil && err != io.EOF {
			log.Printf("error while streaming: %v", err)
		}
	case <-playContext.Done():
		log.Printf("Skipped")
		voiceChat.Skip = nil
		return nil
	}
	voiceChat.Skip = nil
	log.Printf("End of song")
	//log.Printf("%s", encodeSession.FFMPEGMessages())
	return nil
}
