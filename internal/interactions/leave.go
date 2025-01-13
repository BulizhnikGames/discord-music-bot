package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func LeaveInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.LeaveVoiceChat(interaction.GuildID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, fmt.Sprintf("⛔  left voice channel  ⛔"))
		go func() {
			_ = bot.SetPlaybackMessageToText(interaction.GuildID, "⛔  left voice channel  ⛔")
		}()
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
