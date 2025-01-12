package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
	ytsearch "github.com/kkdai/youtube/v2"
	"log"
	"strings"
	"time"
)

func PlayInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		return play(bot, interaction)
	case discordgo.InteractionApplicationCommandAutocomplete:
		/*err := autoComplete(bot, interaction)
		//dont handle error if it's 403 - quota exited
		if err != nil && strings.Contains(err.Error(), "403") {
			err = nil
		}
		return err*/
		_ = autoComplete(bot, interaction)
		return nil
	default:
		return errors.Newf("unknown interaction type: %s", interaction.Type.String())
	}
}

func play(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	data := interaction.ApplicationCommandData()
	song := data.Options[0].StringValue()

	var videos []*ytsearch.PlaylistEntry
	if strings.Contains(song, config.LINK_PREFIX) {
		client := ytsearch.Client{}
		playlist, err := client.GetPlaylist(song)
		if err == nil {
			if len(playlist.Videos) > config.QUEUE_SIZE {
				return errors.New("playlist is too big").AddUser("couldn't playlist to queue, because of it's too big")
			}
			videos = make([]*ytsearch.PlaylistEntry, len(playlist.Videos))
			for i, video := range playlist.Videos {
				videos[i] = video
			}
		}
	}

	channelID, err := bot.GetUsersVoiceChat(interaction.GuildID, interaction.Member.User)
	if err != nil {
		return errors.Newf("%v", err).AddUser("you must be in voice chat")
	}

	voiceChat, err := bot.JoinVoiceChat(interaction.GuildID, channelID, interaction.ChannelID)
	if err != nil {
		return err
	}

	if len(videos) == 0 {
		voiceChat.InsertQueue(internal.Song{Query: song})
	} else {
		go func() {
			log.Printf("!")
			for _, video := range videos {
				voiceChat.InsertQueue(internal.Song{
					Title:    video.Title,
					Author:   video.Author,
					Query:    config.LINK_PREFIX + video.ID,
					Duration: int(video.Duration / time.Second),
				})
			}
		}()
	}

	var message string
	if len(videos) == 0 {
		message = fmt.Sprintf("Added song to queue: %s", song)
	} else {
		message = fmt.Sprintf("Added playlist to queue: %s", song)
	}
	responseToInteraction(bot, interaction, message)

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
			return errors.Newf("Error getting YT videos by with name %s: %s \n", input, err)
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
		return errors.Newf("couldn't send play autocomplete options (%v) to user: %v", choices, err)
	}

	return nil
}
