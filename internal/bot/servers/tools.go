package servers

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"strings"
)

func (server *Server) SetDJRole(id string) error {
	return server.db.Set(context.Background(), "dj:"+server.GuildID, id, 0).Err()
}

func (server *Server) DeleteDJRole() error {
	return server.db.Del(context.Background(), "dj:"+server.GuildID).Err()
}

// GetDJRole returns DJ role's id and true if servers has one, empty string and false - otherwise
func (server *Server) GetDJRole() (string, bool, error) {
	res, err := server.db.Get(context.Background(), "dj:"+server.GuildID).Result()
	if err != nil {
		if err == redis.Nil || strings.Contains(err.Error(), "No connection could be made") {
			return "", false, nil
		}
		return "", false, err
	}
	return res, true, nil
}

func (server *Server) GetUsersVoiceChat(user *discordgo.User) (string, error) {
	guild, err := server.Session.State.Guild(server.GuildID)
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

func (server *Server) TryRegenPlaybackMessage() {
	go server.VoiceChat.TryRegenPlaybackMessage(server.Session)
}

func (server *Server) SetPlaybackMessageToText(text string) error {
	if server.VoiceChat != nil {
		return server.VoiceChat.SetPlaybackMessageToText(server.Session, text)
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}
