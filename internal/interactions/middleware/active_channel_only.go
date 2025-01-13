package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

func ActiveChannelOnly(next bot.InteractionFunc, mustBeInChannel any) bot.InteractionFunc {
	return func(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
		if interaction.Type == discordgo.InteractionApplicationCommand || interaction.Type == discordgo.InteractionMessageComponent {
			guildID := interaction.GuildID
			bot.VoiceEntities.Mutex.RLock()
			voiceChat, ok := bot.VoiceEntities.Data[guildID]
			bot.VoiceEntities.Mutex.RUnlock()
			if ok {
				//log.Printf("voiceChat: %v", voiceChat)
				channelID, err := bot.GetUsersVoiceChat(guildID, interaction.Member.User)
				if err != nil {
					return errors.Newf("couldn't get users's voice chat for middleware: %v", err)
				}
				if channelID == voiceChat.GetVoiceChatID() {
					return next(bot, interaction)
				} else {
					return errors.New("user is not in bot's active voice chat").AddUser("you must be in channel with bot")
				}
			} else {
				switch arg := mustBeInChannel.(type) {
				case bool:
					if arg {
						return errors.Newf("bot is not in voice chat").AddUser("bot must be in voice chat")
					} else {
						return next(bot, interaction)
					}
				default:
					return errors.New("invalid middleware argument argument")
				}
			}
		} else {
			return next(bot, interaction)
		}
	}
}
