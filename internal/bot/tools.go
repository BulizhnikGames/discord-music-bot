package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"github.com/bwmarrin/discordgo"
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
