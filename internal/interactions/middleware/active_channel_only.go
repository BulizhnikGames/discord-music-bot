package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func ActiveChannelOnly(next bot.InteractionFunc, mustBeInChannel any) bot.InteractionFunc {
	return func(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
		guildID := interaction.GuildID
		bot.VoiceEntities.Mutex.RLock()
		voiceChat, ok := bot.VoiceEntities.Data[guildID]
		bot.VoiceEntities.Mutex.RUnlock()
		if ok {
			//log.Printf("voiceChat: %v", voiceChat)
			channelID, err := bot.GetUsersVoiceChat(guildID, interaction.Member.User)
			if err != nil {
				return errors.Errorf("couldn't get users's voice chat for middleware: %v", err)
			}
			if channelID == voiceChat.GetVoiceChatID() {
				return next(bot, interaction)
			} else {
				return errors.New("user is not in bot's active voice chat")
			}
		} else {
			switch arg := mustBeInChannel.(type) {
			case bool:
				if arg {
					return errors.Errorf("bot is not in voice chat")
				} else {
					return next(bot, interaction)
				}
			default:
				return errors.Errorf("invalid middleware argument argument")
			}
		}
	}
}
