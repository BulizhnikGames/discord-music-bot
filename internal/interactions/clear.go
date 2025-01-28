package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func ClearInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.ClearQueue("")
		if err != nil {
			return err
		}
		responseToInteraction(server, interaction, "⏹️  queue cleared  ⏹️")
		return nil
	case discordgo.InteractionMessageComponent:
		return server.ClearQueue("⏹️  queue cleared  ⏹️")
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
