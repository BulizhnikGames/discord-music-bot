package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func LoopInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		loop := interaction.ApplicationCommandData().Options[0].IntValue()
		err := server.SetLoop(int(loop))
		if err != nil {
			return err
		}
		switch loop {
		case 0:
			responseToInteraction(server, interaction, "↪️  looping disabled  ↪️")
		case 1:
			responseToInteraction(server, interaction, "🔁  looping over queue  🔁")
		case 2:
			responseToInteraction(server, interaction, "🔂  looping over song  🔂")
		default:
			responseToInteraction(server, interaction, "↪️  looping disabled  ↪️")
		}
		go server.TryRegenPlaybackMessage()
		return nil
	case discordgo.InteractionApplicationCommandAutocomplete:
		err := server.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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

func Loop0(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionMessageComponent:
		err := server.SetLoop(0)
		if err != nil {
			return err
		}
		server.TryRegenPlaybackMessage()
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func Loop1(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionMessageComponent:
		err := server.SetLoop(1)
		if err != nil {
			return err
		}
		server.TryRegenPlaybackMessage()
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func Loop2(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionMessageComponent:
		err := server.SetLoop(2)
		if err != nil {
			return err
		}
		server.TryRegenPlaybackMessage()
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
