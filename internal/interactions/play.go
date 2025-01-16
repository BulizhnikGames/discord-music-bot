package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
	ytsearch "github.com/kkdai/youtube/v2"
	"log"
	"strings"
	"time"
)

type searcher interface {
	Search(query string, single bool) ([]string, error)
}

func PlayInteraction(yt searcher) servers.InteractionFunc {
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
			_ = autoComplete(server, yt, interaction)
			return nil
		default:
			return errors.Newf("unknown interaction type: %s", interaction.Type.String())
		}
	}
}

func play(server *servers.Server, interaction *discordgo.InteractionCreate) error {
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
			responseToInteraction(server.Session, interaction, "✅  added to queue  ✅", song)
		} else {
			title := fmt.Sprintf(
				"%s - [%d:%02d]",
				metadata.Title,
				int(metadata.Duration/time.Second)/60,
				int(metadata.Duration/time.Second)%60,
			)
			if len(metadata.Thumbnails) == 0 {
				responseToInteraction(
					server.Session,
					interaction,
					"✅  added to queue  ✅",
					title,
					song,
					metadata.Author,
				)
			} else {
				responseToInteraction(
					server.Session,
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
		var title = song
		if playlist.Title != "" {
			title = playlist.Title
		}
		if len(videos[0].Thumbnails) == 0 {
			responseToInteraction(
				server.Session,
				interaction,
				"✅  added to queue  ✅",
				title,
				song,
				playlist.Author,
			)
		} else {
			responseToInteraction(
				server.Session,
				interaction,
				"✅  added to queue  ✅",
				title,
				song,
				playlist.Author,
				videos[0].Thumbnails[0].URL,
			)
		}
	}

	return nil
}

func autoComplete(server *servers.Server, yt searcher, interaction *discordgo.InteractionCreate) error {
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
		results, err := yt.Search(input, false)
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
