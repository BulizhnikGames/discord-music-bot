package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func SkipInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.SkipSong("")
		if err != nil {
			return err
		}
		responseToInteraction(server.Session, interaction, "⏩ skipped")
		return nil
	case discordgo.InteractionMessageComponent:
		return server.SkipSong("⏩ skipped")
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
