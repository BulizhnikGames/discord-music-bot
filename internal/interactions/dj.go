package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

func DJModeInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		guildID := interaction.GuildID
		role := interaction.ApplicationCommandData().Options[0].RoleValue(server.Session, guildID)
		err := server.SetDJRole(role.ID)
		if err != nil {
			return err
		}
		responseToInteraction(server.Session, interaction, fmt.Sprintf("🥁  DJ role is set to <@&%s>  🥁", role.ID))
		return nil
	case discordgo.InteractionApplicationCommandAutocomplete:
		guildID := interaction.GuildID
		guild, err := server.Session.State.Guild(guildID)
		if err != nil {
			return errors.Newf("couldn't get guild with id %s", guildID)
		}
		choices := make([]*discordgo.ApplicationCommandOptionChoice, len(guild.Roles)+1)
		choices[0] = &discordgo.ApplicationCommandOptionChoice{
			Name:  "none",
			Value: &discordgo.Role{},
		}
		for i, role := range guild.Roles {
			choices[i+1] = &discordgo.ApplicationCommandOptionChoice{
				Name:  role.Name,
				Value: role,
			}
		}
		err = server.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		})
		if err != nil {
			return errors.Newf("couldn't send dj-role autocomplete options to user: %v", err)
		}
		return nil
	default:
		return errors.Newf("unknown interaction type: %s", interaction.Type.String())
	}
}

func NoDJInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.DeleteDJRole()
		if err != nil {
			return err
		}
		responseToInteraction(server.Session, interaction, fmt.Sprintf("🥁  DJ role unseted  🥁"))
		return nil
	default:
		return errors.Newf("unknown interaction type: %s", interaction.Type.String())
	}
}
