package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers/voice"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"log"
	"strconv"
	"strings"
)

const PAGE_SIZE = 20

func QueueInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		queue, err := server.GetQueue()
		if err != nil {
			return err
		}
		if len(queue) == 0 {
			responseToInteraction(server.Session, interaction, "📻  playback queue is empty  📻")
			return nil
		}
		respWithQueuePage(server.Session, interaction, queue, 0)
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
		respWithQueuePage(server.Session, interaction, queue, page+1)
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
		respWithQueuePage(server.Session, interaction, queue, page-1)
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

func respWithQueuePage(session *discordgo.Session, inter *discordgo.InteractionCreate, queue []string, page int) {
	log.Printf("queue page: %d", page)
	if page < 0 {
		return
	}
	if len(queue) <= (page)*PAGE_SIZE {
		if len(queue) == 0 {
			responseToInteraction(session, inter, "📻  playback queue is empty  📻")
		} else {
			respWithQueuePage(session, inter, queue, (len(queue)-1)/PAGE_SIZE)
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
	respToQueueInter(session, inter, page > 0, len(queue) > (page+1)*PAGE_SIZE, resp...)
}

func respToQueueInter(sess *discordgo.Session, inter *discordgo.InteractionCreate, prev, next bool, elems ...string) {
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
          "custom_id": "`+sess.State.SessionID+`:queueprev",
          "type": 2,
          "style": 2,
          "label": "prev page"
        }`)
	}
	if next {
		rawComps = append(rawComps, `{
          "custom_id": "`+sess.State.SessionID+`:queuenext",
          "type": 2,
          "style": 2,
          "label": "next page"
        }`)
	}

	line, err := discordgo.MessageComponentFromJSON(voice.ConstructJsonLine(rawComps...))
	if err != nil {
		return
	}

	_, err = sess.InteractionResponseEdit(inter.Interaction, &discordgo.WebhookEdit{
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
		Components: &[]discordgo.MessageComponent{
			line,
		},
	})
	if err != nil {
		log.Printf("Failed to respond to queue interaction: %v", err)
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
				server.Session,
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
				song.FileUrl,
				fmt.Sprintf("by %s", song.Author),
				song.ThumbnailUrl,
			)
		} else {
			responseToInteraction(server.Session, interaction, "🔇  nothing is playing  🔇")
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
