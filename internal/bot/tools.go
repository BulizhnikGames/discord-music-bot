package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"log"
)

func (bot *DiscordBot) GetUsersVoiceChat(guildID string, user *discordgo.User) (string, error) {
	guild, err := bot.Session.State.Guild(guildID)
	if err != nil {
		return "", err
	}
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == user.ID {
			return voiceState.ChannelID, nil
		}
	}
	return "", errors.New("couldn't get user's voice chat")
}

func (bot *DiscordBot) SendInChannel(channelID, message string) error {
	_, err := bot.Session.ChannelMessageSend(channelID, message)
	return err
}

func (bot *DiscordBot) SetDJRole(guildID, id string) error {
	return bot.db.Set(context.Background(), "dj:"+guildID, id, 0).Err()
}

func (bot *DiscordBot) DeleteDJRole(guildID string) error {
	return bot.db.Del(context.Background(), "dj:"+guildID).Err()
}

// GetDJRole returns DJ role's id and true if server has one, empty string and false - otherwise
func (bot *DiscordBot) GetDJRole(guildID string) (string, bool, error) {
	res, err := bot.db.Get(context.Background(), "dj:"+guildID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return res, true, nil
}

func (voiceChat *VoiceEntity) downloadSong(ctx context.Context, song *internal.Song) (*internal.Song, error) {
	voiceChat.cache.Mutex.Lock()
	if cache, ok := voiceChat.cache.Data[song.Query]; ok {
		log.Printf("Song is already in cache: %s", song.Query)
		cache.Cnt++
		voiceChat.cache.Mutex.Unlock()
		return cache.Song, nil
	}
	voiceChat.cache.Mutex.Unlock()
	log.Printf("Downloading song: %s", song.Query)
	res := make(chan youtube.Result)
	go youtube.Download(ctx, voiceChat.voiceConnection.GuildID, song.Query, res)
	select {
	case data := <-res:
		if data.Err != nil {
			return nil, errors.Newf("couldn't download song: %v", data.Err)
		}
		log.Printf("Downloaded song: %v", data.Downloaded.Title)
		voiceChat.cache.Mutex.Lock()
		defer voiceChat.cache.Mutex.Unlock()
		voiceChat.cache.Data[song.Query] = &internal.SongCache{
			Cnt:  1,
			Song: data.Downloaded,
		}
		return data.Downloaded, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (voiceChat *VoiceEntity) GetVoiceChatID() string {
	return voiceChat.voiceConnection.ChannelID
}

func (bot *DiscordBot) TryRegenPlaybackMessage(guildID string) {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		go voiceChat.TryRegenPlaybackMessage(bot.Session)
	}
}

func (bot *DiscordBot) SetPlaybackMessageToText(guildID, text string) error {
	bot.VoiceEntities.Mutex.RLock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		bot.VoiceEntities.Mutex.RUnlock()
		return voiceChat.SetPlaybackMessageToText(bot.Session, text)
	} else {
		bot.VoiceEntities.Mutex.RUnlock()
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}
