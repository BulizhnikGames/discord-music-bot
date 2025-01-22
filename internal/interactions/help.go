package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func HelpInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		_, err := server.Session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    "🗒️  information  🗒️",
						IconURL: interaction.Member.User.AvatarURL("64x64"),
					},
					Title: "list of commands",
					Color: 2326507,
					Fields: []*discordgo.MessageEmbedField{
						{Name: "/clear - clear playback queue and stop playback"},
						{Name: "/dj-mode - set dj mode, only members with dj role can manipulate queue and playback"},
						{Name: "/dj-off - turn off dj mode"},
						{Name: "/help - get list of commands and their meanings"},
						{Name: "/leave - leave voice chat"},
						{Name: "/loop - loop playback (0 - no loop, 1 - loop over queue, 2 - loop over song)"},
						{Name: "/nowplaying - get information about current song"},
						{Name: "/pause - pause playback"},
						{Name: "/play - play song by youtube link or query"},
						{Name: "/queue - get playback queue"},
						{Name: "/resume - resume playback"},
						{Name: "/shuffle - shuffle playback queue"},
						{Name: "/skip - skip current song"},
						{Name: "/stop - clear playback queue and stop playback"},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "github.com/BulizhnikGames/discord-music-bot",
					},
				},
			},
		})
		return err
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
