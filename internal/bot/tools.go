package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
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

func (bot *DiscordBot) StopPlayback(guildID string) error {
	bot.VoiceEntities.Mutex.Lock()
	defer bot.VoiceEntities.Mutex.Unlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.stop()
		return nil
	} else {
		return errors.Errorf("bot doesn't have active playback in guild %s", guildID)
	}
}

func (bot *DiscordBot) SendInChannel(channelID, message string) error {
	_, err := bot.Session.ChannelMessageSend(channelID, message)
	return err
}

func (voiceChat *VoiceEntity) downloadSong(ctx context.Context, song *internal.Song) *internal.Song {
	voiceChat.cache.Mutex.Lock()
	defer voiceChat.cache.Mutex.Unlock()
	if cache, ok := voiceChat.cache.Data[song.Query]; ok {
		log.Printf("Song is already in cache: %s", song.Query)
		cache.Cnt++
		return cache.Song
	} else {
		log.Printf("Downloading song: %s", song.Query)
		downloaded, err := youtube.Download(ctx, voiceChat.voiceConnection.GuildID, song.Query)
		if err != nil {
			log.Printf("Couldn't download song: %v", err)
			return &internal.Song{
				Query: song.Query,
			}
		}
		log.Printf("Downloaded song: %v", downloaded.Title)
		cache = &internal.SongCache{
			Cnt:  1,
			Song: downloaded,
		}
		voiceChat.cache.Data[song.Query] = cache
		return downloaded
	}
}

func (voiceChat *VoiceEntity) clearCacheCnt() {
	voiceChat.cache.Mutex.Lock()
	defer voiceChat.cache.Mutex.Unlock()
	for _, cache := range voiceChat.cache.Data {
		cache.Cnt = 0
	}
}
