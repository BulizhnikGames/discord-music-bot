package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
)

func DJModeInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		guildID := interaction.GuildID
		role := interaction.ApplicationCommandData().Options[0].RoleValue(bot.Session, guildID)
		err := bot.SetDJRole(guildID, role.ID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, fmt.Sprintf(":drum:  DJ role is set to <@&%s>  :drum:", role.ID))
		return nil
	case discordgo.InteractionApplicationCommandAutocomplete:
		guildID := interaction.GuildID
		guild, err := bot.Session.State.Guild(guildID)
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
		err = bot.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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

func NoDJInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		guildID := interaction.GuildID
		err := bot.DeleteDJRole(guildID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, fmt.Sprintf(":drum:  DJ role unseted  :drum:"))
		return nil
	default:
		return errors.Newf("unknown interaction type: %s", interaction.Type.String())
	}
}
