package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
	"log"
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
		responseToDJInteraction(
			server.Session,
			interaction,
			"🥁  DJ role is set  🥁",
			fmt.Sprintf("You must have <@&%s> role to manipulate playback", role.ID),
		)
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
		responseToDJInteraction(
			server.Session,
			interaction,
			"🥁  DJ role unset  🥁",
			"Now you don't need specific role to manipulate playback",
		)
		return nil
	default:
		return errors.Newf("unknown interaction type: %s", interaction.Type.String())
	}
}

func responseToDJInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate, elems ...string) {
	for len(elems) < 2 {
		elems = append(elems, "")
	}
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Author: &discordgo.MessageEmbedAuthor{
					Name:    elems[0],
					IconURL: interaction.Member.User.AvatarURL("64x64"),
				},
				Description: elems[1],
				Color:       2326507,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "github.com/BulizhnikGames/discord-music-bot",
				},
			},
		},
	})
	if err != nil {
		log.Printf("Failed to respond to interaction (with text %s): %v", elems[0], err)
	}
}
