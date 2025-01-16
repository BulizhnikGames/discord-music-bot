package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

const AdminPermissions = discordgo.PermissionAdministrator

func AdminOnly(next servers.InteractionFunc) servers.InteractionFunc {
	return func(server *servers.Server, interaction *discordgo.InteractionCreate) error {
		if interaction.Type == discordgo.InteractionApplicationCommand {
			if interaction.Member.Permissions&AdminPermissions == AdminPermissions {
				return next(server, interaction)
			} else {
				return errors.New("not allowed").AddUser("you must have admin permissions to do this")
			}
		} else {
			return next(server, interaction)
		}
	}
}
