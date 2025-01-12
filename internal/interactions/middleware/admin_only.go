package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

const AdminPermissions = discordgo.PermissionAdministrator

func AdminOnly(next bot.InteractionFunc) bot.InteractionFunc {
	return func(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
		if interaction.Type == discordgo.InteractionApplicationCommand {
			if interaction.Member.Permissions&AdminPermissions == AdminPermissions {
				return next(bot, interaction)
			} else {
				return errors.New("not allowed").AddUser("you must have admin permissions to do this")
			}
		} else {
			return next(bot, interaction)
		}
	}
}
