package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func LoopInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		loop := interaction.ApplicationCommandData().Options[0].IntValue()
		err := bot.SetLoop(interaction.GuildID, int(loop))
		if err != nil {
			return err
		}
		switch loop {
		case 0:
			responseToInteraction(bot, interaction, fmt.Sprintf(":ballot_box_with_check: looping disabled :ballot_box_with_check:"))
		case 1:
			responseToInteraction(bot, interaction, fmt.Sprintf(":repeat: looping over queue :repeat:"))
		case 2:
			responseToInteraction(bot, interaction, fmt.Sprintf(":repeat_one: looping over song :repeat_one:"))
		default:
			responseToInteraction(bot, interaction, fmt.Sprintf(":ballot_box_with_check: looping disabled :ballot_box_with_check:"))
		}
		return nil
	case discordgo.InteractionApplicationCommandAutocomplete:
		err := bot.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "no loop",
						Value: 0,
					},
					{
						Name:  "loop queue",
						Value: 1,
					},
					{
						Name:  "loop song",
						Value: 2,
					},
				},
			},
		})
		if err != nil {
			return errors.Errorf("couldn't send loop autocomplete options to user: %v", err)
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
