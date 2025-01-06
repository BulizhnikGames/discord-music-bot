package commands

import (
	"errors"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube/api"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func Play(s *api.Youtube, session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		data := interaction.ApplicationCommandData()
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: data.Options[0].StringValue(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("Errro responding to interaction: %s \n", err)
			return
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := interaction.ApplicationCommandData()
		input := data.Options[0].StringValue()
		if input == "" {
			return
		}
		choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
		if strings.HasPrefix(input, "http") || strings.HasPrefix("http", input) {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  input,
				Value: input,
			})
		} else {
			results, err := getVideos(s, input, 5)
			if err != nil {
				log.Printf("Error getting YT videos by with name %s: %s \n", input, err)
				return
			}
			//log.Printf("Got %v names from search", len(*results))
			if len(*results) == 0 {
				return
			}
			for _, result := range *results {
				lists := strings.Split(result, " ")
				name := ""
				for i, part := range lists {
					if i == 0 {
						continue
					}
					name += part
					if i != len(lists)-1 {
						name += " "
					}
				}
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  name,
					Value: lists[0],
				})
			}
		}

		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		})
		if err != nil {
			log.Printf("Errro responding to interaction: %s \n", err)
			return
		}
	}
}

func getVideos(s *api.Youtube, input string, amountLimit int) (*[]string, error) {
	if input == "" {
		return nil, errors.New("input is empty")
	}

	names, err := s.Search(input, amountLimit)
	return names, err
}
