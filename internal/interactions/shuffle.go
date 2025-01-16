package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func ShuffleInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.ShuffleQueue()
		if err != nil {
			return err
		}
		responseToInteraction(server.Session, interaction, "🔀  shuffled  🔀")
		return nil
	case discordgo.InteractionMessageComponent:
		err := server.ShuffleQueue()
		if err != nil {
			return err
		}
		_, err = server.Session.ChannelMessageSendEmbed(interaction.ChannelID, &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "🔀  shuffled  🔀",
				IconURL: server.Session.State.User.AvatarURL("64x64"),
			},
			Color: 2326507,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "github.com/BulizhnikGames/discord-music-bot",
			},
		})
		return err
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
