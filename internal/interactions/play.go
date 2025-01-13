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

	var playlist *ytsearch.Playlist
	var videos []*ytsearch.PlaylistEntry
	var metadata *ytsearch.Video
	var err error
	if strings.Contains(song, config.LINK_PREFIX) {
		client := ytsearch.Client{}
		playlist, err = client.GetPlaylist(song)
		if err == nil {
			if len(playlist.Videos) > config.QUEUE_SIZE {
				return errors.New("playlist is too big").AddUser("couldn't playlist to queue, because of it's too big")
			}
			videos = make([]*ytsearch.PlaylistEntry, len(playlist.Videos))
			for i, video := range playlist.Videos {
				videos[i] = video
			}
		} else {
			metadata, err = client.GetVideo(song)
			if err != nil {
				log.Printf("Error getting video metadata: %s", err)
			}
		}
	}

	channelID, err := bot.GetUsersVoiceChat(interaction.GuildID, interaction.Member.User)
	if err != nil {
		return errors.Newf("%v", err).AddUser("you must be in voice chat")
	}

	//log.Printf("getting voice chat")
	voiceChat, err := bot.JoinVoiceChat(interaction.GuildID, channelID, interaction.ChannelID)
	if err != nil {
		return err
	}
	//log.Printf("got voice chat")

	if len(videos) == 0 {
		if metadata == nil {
			voiceChat.InsertQueue(internal.Song{Query: song})
		} else {
			voiceChat.InsertQueue(internal.Song{
				Title:    metadata.Title,
				Author:   metadata.Author,
				Query:    song,
				Duration: int(metadata.Duration / time.Second),
			})
		}
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

	if len(videos) == 0 {
		if metadata == nil {
			responseToInteraction(bot, interaction, "✅  added to queue  ✅", song)
		} else {
			title := fmt.Sprintf(
				"%s - [%d:%02d]",
				metadata.Title,
				int(metadata.Duration/time.Second)/60,
				int(metadata.Duration/time.Second)%60,
			)
			if len(metadata.Thumbnails) == 0 {
				responseToInteraction(
					bot,
					interaction,
					"✅  added to queue  ✅",
					title,
					song,
					metadata.Author,
				)
			} else {
				responseToInteraction(
					bot,
					interaction,
					"✅  added to queue  ✅",
					title,
					song,
					metadata.Author,
					metadata.Thumbnails[0].URL,
				)
			}
		}
	} else {
		if len(videos[0].Thumbnails) == 0 {
			responseToInteraction(
				bot,
				interaction,
				"✅  added to queue  ✅",
				playlist.Title,
				song,
				playlist.Author,
			)
		} else {
			responseToInteraction(
				bot,
				interaction,
				"✅  added to queue  ✅",
				playlist.Title,
				song,
				playlist.Author,
				videos[0].Thumbnails[0].URL,
			)
		}
	}

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
