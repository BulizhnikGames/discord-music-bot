package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
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
		voiceChat.Stop()
		return nil
	} else {
		return errors.Errorf("bot doesn't have active playback in guild %s", guildID)
	}
}
