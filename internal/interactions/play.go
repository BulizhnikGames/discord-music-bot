package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"strings"
)

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

	channelID, err := bot.GetUsersVoiceChat(interaction.GuildID, interaction.Member.User)
	if err != nil {
		return errors.Wrap(err, "User is not in voice chat")
	}

	queue, err := bot.JoinVoiceChat(interaction.GuildID, channelID, interaction.ChannelID)
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

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	if strings.HasPrefix(input, config.LINK_PREFIX) {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  input,
			Value: input,
		})
	} else {
		results, err := bot.Youtube.Search(input, false)
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
				Value: config.LINK_PREFIX + result[:idx],
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
