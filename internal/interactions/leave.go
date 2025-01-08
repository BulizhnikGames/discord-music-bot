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
		err := bot.StopPlayback(interaction.GuildID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, fmt.Sprintf(":no_entry: left voice channel :no_entry:"), 0)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
