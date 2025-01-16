package middleware

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

func DJOrAdminOnly(next servers.InteractionFunc) servers.InteractionFunc {
	return func(server *servers.Server, interaction *discordgo.InteractionCreate) error {
		if interaction.Type == discordgo.InteractionApplicationCommand || interaction.Type == discordgo.InteractionMessageComponent {
			if interaction.Member.Permissions&AdminPermissions == AdminPermissions {
				return next(server, interaction)
			} else {
				djRole, have, err := server.GetDJRole()
				if err != nil {
					return err
				}
				if !have {
					return next(server, interaction)
				}
				flag := false
				for _, role := range interaction.Member.Roles {
					if role == djRole {
						flag = true
					}
				}
				if flag {
					return next(server, interaction)
				}
				return errors.New("not allowed").AddUser("you must have dj role to do that")
			}
		} else {
			return next(server, interaction)
		}
	}
}
