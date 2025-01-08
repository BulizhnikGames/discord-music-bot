package bot

import (
	"context"
	"github.com/go-faster/errors"
)

// JoinVoiceChat bot joins voice chat if it isn't already in one.
// If bot is already in voice chat:
// 1. Returns error if he's in other voice chat
// 2. Returns queue if he's in requested voice chat
func (bot *DiscordBot) JoinVoiceChat(guildID string, channelID string) (chan string, error) {
	bot.VoiceEntities.Mutex.RLock()
	if v, ok := bot.VoiceEntities.Data[guildID]; ok {
		if v.VoiceConnection.ChannelID == channelID {
			bot.VoiceEntities.Mutex.RUnlock()
			return v.Queue, nil
		} else {
			err := errors.Errorf("could not join VoiceConnection chat (id: %s, guild: %s): already in another channel", v.VoiceConnection.ChannelID, guildID)
			bot.VoiceEntities.Mutex.RUnlock()
			return nil, errors.Wrap(err, "Bot is already in other channel")
		}
	}
	bot.VoiceEntities.Mutex.RUnlock()

	voiceData, err := bot.Session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		err = errors.Errorf("could not join VoiceConnection chat (id: %s, guild: %s): %v", channelID, guildID, err)
		return nil, errors.Wrap(err, "Couldn't join voice chat")
	}

	queue := make(chan string, 200)
	voiceChat := &VoiceEntity{
		VoiceConnection: voiceData,
		Queue:           queue,
	}
	var ctx context.Context
	ctx, voiceChat.Stop = context.WithCancel(context.Background())

	bot.VoiceEntities.Put(guildID, voiceChat)

	go voiceChat.PlaySongs(ctx, bot)

	return voiceChat.Queue, nil
}
