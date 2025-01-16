package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

func ActiveChannelOnly(next servers.InteractionFunc, mustBeInChannel any) servers.InteractionFunc {
	return func(server *servers.Server, interaction *discordgo.InteractionCreate) error {
		if interaction.Type == discordgo.InteractionApplicationCommand || interaction.Type == discordgo.InteractionMessageComponent {
			if server.VoiceChat != nil {
				//log.Printf("voiceChat: %v", voiceChat)
				channelID, err := server.GetUsersVoiceChat(interaction.Member.User)
				if err != nil {
					return errors.Newf("couldn't get users's voice chat for middleware: %v", err)
				}
				if channelID == server.VoiceChat.GetVoiceChatID() {
					return next(server, interaction)
				} else {
					return errors.New("user is not in bot's active voice chat").AddUser("you must be in channel with bot")
				}
			} else {
				switch arg := mustBeInChannel.(type) {
				case bool:
					if arg {
						return errors.Newf("bot is not in voice chat").AddUser("bot must be in voice chat")
					} else {
						return next(server, interaction)
					}
				default:
					return errors.New("invalid middleware argument argument")
				}
			}
		} else {
			return next(server, interaction)
		}
	}
}
