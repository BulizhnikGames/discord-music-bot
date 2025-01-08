package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"strings"
)

const SEARCH_LIMIT = 5
const LINK_PREFIX = "https://www.youtube.com/watch?v="

func PlayInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		return play(bot, interaction)
	case discordgo.InteractionApplicationCommandAutocomplete:
		return autoComplete(bot, interaction)
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func play(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	data := interaction.ApplicationCommandData()
	song := data.Options[0].StringValue()
	if !strings.HasPrefix(song, LINK_PREFIX) {
		results, err := bot.Youtube.Search(song, 1)
		if err != nil {
			return errors.Wrap(err, "Couldn't get song")
		}
		if len(results) == 0 {
			err = errors.New("YT videos not found")
			return errors.Wrap(err, "Couldn't get song")
		}
		idx := strings.Index(results[0], ":")
		song = LINK_PREFIX + results[0][:idx]
	}

	channelID, err := bot.GetUsersVoiceChat(interaction.GuildID, interaction.Member.User)
	if err != nil {
		return errors.Wrap(err, "User is not in voice chat")
	}

	queue, err := bot.JoinVoiceChat(interaction.GuildID, channelID)
	if err != nil {
		return err
	}
	queue <- song

	responseToInteraction(bot, interaction, fmt.Sprintf("Added to queue: %s", song), discordgo.MessageFlagsEphemeral)

	return nil
}

func autoComplete(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	data := interaction.ApplicationCommandData()
	input := data.Options[0].StringValue()
	if input == "" {
		return nil
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, SEARCH_LIMIT)
	if strings.HasPrefix(input, LINK_PREFIX) {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  input,
			Value: input,
		})
	} else {
		results, err := bot.Youtube.Search(input, SEARCH_LIMIT)
		if err != nil {
			return errors.Errorf("Error getting YT videos by with name %s: %s \n", input, err)
		}
		//log.Printf("Got %v names from search", len(*results))
		if len(results) == 0 {
			return errors.New("YT videos not found")
		}

		for _, result := range results {
			idx := strings.Index(result, ":")
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  result[idx+1:],
				Value: LINK_PREFIX + result[:idx],
			})
		}
	}

	err := bot.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		return errors.Errorf("error trying to send autocomplete options to user: %s", err)
	}

	return nil
}
