package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/go-faster/errors"
	"sync"
)

// JoinVoiceChat bot joins voice chat if it isn't already in one.
// If bot is already in voice chat:
// 1. Returns error if he's in other voice chat
// 2. Returns queue if he's in requested voice chat
func (bot *DiscordBot) JoinVoiceChat(guildID, voiceChannel, textChannel string) (*VoiceEntity, error) {
	bot.VoiceEntities.Mutex.RLock()
	if v, ok := bot.VoiceEntities.Data[guildID]; ok {
		if v.voiceConnection.ChannelID == voiceChannel {
			bot.VoiceEntities.Mutex.RUnlock()
			return v, nil
		} else {
			err := errors.Errorf("could not join voiceConnection chat (id: %s, guild: %s): already in another channel", v.voiceConnection.ChannelID, guildID)
			bot.VoiceEntities.Mutex.RUnlock()
			return nil, errors.Wrap(err, "Bot is already in other channel")
		}
	}
	bot.VoiceEntities.Mutex.RUnlock()

	voiceData, err := bot.Session.ChannelVoiceJoin(guildID, voiceChannel, false, true)
	if err != nil {
		err = errors.Errorf("could not join voiceConnection chat (id: %s, guild: %s): %v", voiceChannel, guildID, err)
		return nil, errors.Wrap(err, "Couldn't join voice chat")
	}

	queue := internal.CreateCycleQueue[internal.Song](200)
	voiceChat := &VoiceEntity{
		voiceConnection: voiceData,
		Queue:           queue,
		textChannel:     textChannel,
		cache: internal.AsyncMap[string, *internal.SongCache]{
			Data:  make(map[string]*internal.SongCache),
			Mutex: &sync.RWMutex{},
		},
	}
	queue.SetHandler(voiceChat.downloadSong)
	var ctx context.Context
	ctx, voiceChat.stop = context.WithCancel(context.Background())

	bot.VoiceEntities.Put(guildID, voiceChat)

	go voiceChat.PlaySongs(ctx, bot)

	return voiceChat, nil
}

func (bot *DiscordBot) LeaveVoiceChat(guildID string) error {
	bot.VoiceEntities.Mutex.Lock()
	defer bot.VoiceEntities.Mutex.Unlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.stop()
		err := voiceChat.voiceConnection.Disconnect()
		if err != nil {
			return err
		}
		delete(bot.VoiceEntities.Data, guildID)
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}
