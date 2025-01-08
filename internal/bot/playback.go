package bot

import (
	"context"
	"github.com/go-faster/errors"
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
			log.Printf("Playing song %s", song)
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
