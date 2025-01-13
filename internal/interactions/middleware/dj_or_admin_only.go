package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

func DJOrAdminOnly(next bot.InteractionFunc) bot.InteractionFunc {
	return func(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
		if interaction.Type == discordgo.InteractionApplicationCommand || interaction.Type == discordgo.InteractionMessageComponent {
			if interaction.Member.Permissions&AdminPermissions == AdminPermissions {
				return next(bot, interaction)
			} else {
				djRole, have, err := bot.GetDJRole(interaction.GuildID)
				if err != nil {
					return err
				}
				if !have {
					return next(bot, interaction)
				}
				flag := false
				for _, role := range interaction.Member.Roles {
					if role == djRole {
						flag = true
					}
				}
				if flag {
					return next(bot, interaction)
				}
				return errors.New("not allowed").AddUser("you must have dj role to do that")
			}
		} else {
			return next(bot, interaction)
		}
	}
}
