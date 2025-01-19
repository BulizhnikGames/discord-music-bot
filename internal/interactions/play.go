package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"github.com/bwmarrin/discordgo"
	ytsearch "github.com/kkdai/youtube/v2"
	"log"
	"strings"
	"time"
)

const SEARCH_LIMIT = 1

// SearchFunc gets query for YouTube and cnt of search results and returns array of titles, and array of urls
type SearchFunc func(query string, cnt int) ([]string, []string, error)

// PlayInteraction must be initialized with search func which will be used to autocomplete /play command with search res
func PlayInteraction(search SearchFunc) servers.InteractionFunc {
	return func(server *servers.Server, interaction *discordgo.InteractionCreate) error {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			return play(server, interaction)
		case discordgo.InteractionApplicationCommandAutocomplete:
			/*err := autoComplete(bot, interaction)
			//dont handle error if it's 403 - quota exited
			if err != nil && strings.Contains(err.Error(), "403") {
				err = nil
			}
			return err*/
			_ = autoComplete(server, search, interaction)
			return nil
		default:
			return errors.Newf("unknown interaction type: %s", interaction.Type.String())
		}
	}
}

func play(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	data := interaction.ApplicationCommandData()
	query := data.Options[0].StringValue()

	var playlist *ytsearch.Playlist
	var err error
	if strings.Contains(query, config.LINK_PREFIX) {
		client := ytsearch.Client{}
		playlist, err = client.GetPlaylist(query)
		if err == nil {
			if len(playlist.Videos) > config.QUEUE_SIZE {
				return errors.New("playlist is too big").AddUser("couldn't playlist to queue, because of it's too big")
			}
		}
	}
	var song *internal.Song
	if playlist == nil {
		song, err = youtube.GetMetadata(query, interaction.GuildID, false)
		if err != nil || song == nil {
			return errors.Newf("song not found: %v", err).AddUser("couldn't find song")
		}
	}

	channelID, err := server.GetUsersVoiceChat(interaction.Member.User)
	if err != nil {
		return errors.Newf("%v", err).AddUser("you must be in voice chat")
	}

	//log.Printf("getting voice chat")
	voiceChat, err := server.JoinVoiceChat(interaction.GuildID, channelID, interaction.ChannelID)
	if err != nil {
		return err
	}
	//log.Printf("got voice chat")

	if playlist == nil {
		voiceChat.InsertQueue(*song)
	} else {
		go func() {
			for _, video := range playlist.Videos {
				voiceChat.InsertQueue(internal.Song{
					Title:    video.Title,
					Author:   video.Author,
					Query:    config.LINK_PREFIX + video.ID,
					Duration: int(video.Duration / time.Second),
				})
			}
		}()
	}

	if playlist == nil {
		title := fmt.Sprintf(
			"%s - [%d:%02d]",
			song.Title,
			song.Duration/60,
			song.Duration%60,
		)
		responseToInteraction(
			server.Session,
			interaction,
			"✅  added to queue  ✅",
			title,
			song.OriginalUrl,
			song.Author,
			song.ThumbnailUrl,
		)
	} else {
		var title = query
		if playlist.Title != "" {
			title = playlist.Title
		}
		if len(playlist.Videos[0].Thumbnails) == 0 {
			responseToInteraction(
				server.Session,
				interaction,
				"✅  added to queue  ✅",
				title,
				query,
				playlist.Author,
			)
		} else {
			responseToInteraction(
				server.Session,
				interaction,
				"✅  added to queue  ✅",
				title,
				query,
				playlist.Author,
				playlist.Videos[0].Thumbnails[0].URL,
			)
		}
	}

	return nil
}

func autoComplete(server *servers.Server, search SearchFunc, interaction *discordgo.InteractionCreate) error {
	data := interaction.ApplicationCommandData()
	input := data.Options[0].StringValue()
	if input == "" {
		return nil
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, SEARCH_LIMIT)
	if strings.HasPrefix(input, config.LINK_PREFIX) {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  input,
			Value: input,
		})
	} else {
		titles, urls, err := search(input, SEARCH_LIMIT)
		if err != nil {
			return errors.Newf("Error getting YT videos by with name %s: %s \n", input, err)
		}
		log.Printf("Got %v names from search (%s)", len(titles), input)
		if len(titles) == 0 {
			return errors.New("YT videos not found")
		}

		for i := 0; i < min(len(titles), len(urls)); i++ {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  titles[i],
				Value: urls[i],
			})
		}
	}

	err := server.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
