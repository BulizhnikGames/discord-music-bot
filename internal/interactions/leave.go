package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func LeaveInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.LeaveVoiceChat()
		if err != nil {
			return err
		}
		responseToInteraction(server, interaction, fmt.Sprintf("⛔  left voice channel  ⛔"))
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
