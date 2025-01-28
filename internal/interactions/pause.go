package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func PauseInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.Pause(true)
		if err != nil {
			return err
		}
		go server.TryRegenPlaybackMessage()
		responseToInteraction(server, interaction, "⏸️  playback paused  ⏸️")
		return nil
	case discordgo.InteractionMessageComponent:
		err := server.Pause(true)
		if err != nil {
			return err
		}
		go server.TryRegenPlaybackMessage()
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func ResumeInteraction(server *servers.Server, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := server.Pause(false)
		if err != nil {
			return err
		}
		go server.TryRegenPlaybackMessage()
		responseToInteraction(server, interaction, "▶️  playback resumed  ▶️")
		return nil
	case discordgo.InteractionMessageComponent:
		err := server.Pause(false)
		if err != nil {
			return err
		}
		go server.TryRegenPlaybackMessage()
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
