package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers/voice"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"strconv"
	"strings"
)

const PAGE_SIZE = 10

func QueueInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		queue, err := server.GetQueue()
		if err != nil {
			return err
		}
		if len(queue) == 0 {
			responseToInteraction(server, interaction, "📻  playback queue is empty  📻")
			return nil
		}
		respWithQueuePage(server, interaction, queue, 0)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func QueueNextInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionMessageComponent:
		page, err := getPage(interaction)
		if err != nil {
			return err
		}
		queue, err := server.GetQueue()
		if err != nil {
			return err
		}
		respWithQueuePage(server, interaction, queue, page+1)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func QueuePrevInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionMessageComponent:
		page, err := getPage(interaction)
		if err != nil {
			return err
		}
		queue, err := server.GetQueue()
		if err != nil {
			return err
		}
		respWithQueuePage(server, interaction, queue, page-1)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func getPage(interaction *discordgo.InteractionCreate) (int, error) {
	if len(interaction.Message.Embeds) == 0 {
		return 0, errors.New("no embeds on message")
	}
	if len(interaction.Message.Embeds[0].Fields) == 0 {
		return 0, errors.New("no fields message's embed")
	}
	dotIdx := strings.Index(interaction.Message.Embeds[0].Fields[0].Name, ".")
	if dotIdx == 0 {
		return 0, errors.New("incorrect message format")
	}
	firstIdx, err := strconv.Atoi(interaction.Message.Embeds[0].Fields[0].Name[0:dotIdx])
	if err != nil {
		return 0, errors.New("incorrect message format")
	}
	return firstIdx / PAGE_SIZE, nil
}

func respWithQueuePage(server *servers.Server, inter *discordgo.InteractionCreate, queue []string, page int) {
	if page < 0 {
		return
	}
	if len(queue) <= (page)*PAGE_SIZE {
		if len(queue) == 0 {
			responseToInteraction(server, inter, "📻  playback queue is empty  📻")
		} else {
			respWithQueuePage(server, inter, queue, (len(queue)-1)/PAGE_SIZE)
		}
		return
	}
	n := len(queue) - page*PAGE_SIZE
	resp := make([]string, 1, min(n, PAGE_SIZE)+2)
	resp[0] = "📻  playback queue  📻"
	for i := page * PAGE_SIZE; i < min(len(queue), (page+1)*PAGE_SIZE); i++ {
		resp = append(resp, fmt.Sprintf("%d. %s", i+1, queue[i]))
	}
	resp = append(resp, fmt.Sprintf("page - %d/%d", page+1, (len(queue)-1)/PAGE_SIZE+1))
	respToQueueInter(server, inter, page > 0, len(queue) > (page+1)*PAGE_SIZE, resp...)
}

func respToQueueInter(server *servers.Server, inter *discordgo.InteractionCreate, prev, next bool, elems ...string) {
	if len(elems) == 0 {
		return
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(elems)-1)
	for i := 1; i < len(elems); i++ {
		if i != len(elems)-1 {
			fields = append(fields, &discordgo.MessageEmbedField{Name: elems[i]})
		} else { // last element is smaller
			fields = append(fields, &discordgo.MessageEmbedField{Value: elems[i]})
		}
	}

	rawComps := make([]string, 0, 2)
	if prev {
		rawComps = append(rawComps, `{
          "custom_id": "`+server.Session.State.SessionID+`:queueprev",
          "type": 2,
          "style": 2,
          "label": "prev page"
        }`)
	}
	if next {
		rawComps = append(rawComps, `{
          "custom_id": "`+server.Session.State.SessionID+`:queuenext",
          "type": 2,
          "style": 2,
          "label": "next page"
        }`)
	}

	var components []discordgo.MessageComponent
	if len(rawComps) > 0 {
		line, err := discordgo.MessageComponentFromJSON(voice.ConstructJsonLine(rawComps...))
		if err != nil {
			server.Logger.Printf("failed to construct json line to respond to load queue page: %v", err)
			return
		}
		components = append(components, line)
	}

	_, err := server.Session.InteractionResponseEdit(inter.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Author: &discordgo.MessageEmbedAuthor{
					Name:    elems[0],
					IconURL: inter.Member.User.AvatarURL("64x64"),
				},
				Color:  2326507,
				Fields: fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "github.com/BulizhnikGames/discord-music-bot",
				},
			},
		},
		Components: &components,
	})
	if err != nil {
		server.Logger.Printf("Failed to respond to queue interaction: %v", err)
	}
}

func NowPlayingInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		song, curr, err := server.NowPlaying()
		if err != nil {
			return err
		}
		if song != nil {
			responseToInteraction(
				server,
				interaction,
				"📻  now playing  📻",
				fmt.Sprintf(
					"%s - [%d:%02d / %d:%02d]",
					song.Title,
					curr/60,
					curr%60,
					song.Duration/60,
					song.Duration%60,
				),
				song.OriginalUrl,
				fmt.Sprintf("by %s", song.Author),
				song.ThumbnailUrl,
			)
		} else {
			responseToInteraction(server, interaction, "🔇  nothing is playing  🔇")
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
