package bot

import (
	"context"
	"fmt"
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

func (voiceChat *VoiceEntity) SendPlayBackMessage(bot *DiscordBot, song internal.Song) (*discordgo.Message, error) {
	firstLine, err := discordgo.MessageComponentFromJSON(voiceChat.constructJsonLine(
		voiceChat.pauseButtonJson(),
		voiceChat.skipButtonJson(),
		voiceChat.stopButtonJson(),
	))
	if err != nil {
		return nil, err
	}
	secondLine, err := discordgo.MessageComponentFromJSON(voiceChat.constructJsonLine(
		voiceChat.shuffleQueueJson(),
		voiceChat.loopOptsJson(),
	))
	if err != nil {
		return nil, err
	}
	return bot.Session.ChannelMessageSendComplex(voiceChat.textChannel,
		&discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    "Now Playing",
						IconURL: "https://github.com/BulizhnikGames/discord-music-bot/blob/master/icon.png?raw=true",
					},
					Title: fmt.Sprintf("%s - [%d:%02d]", song.Title, song.Duration/60, song.Duration%60),
					URL:   song.URL,
					Color: 2326507,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: song.Author,
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
}
