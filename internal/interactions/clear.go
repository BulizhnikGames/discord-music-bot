package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func ClearInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.ClearQueue(interaction.GuildID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, fmt.Sprintf(":stop_button: queue cleared :stop_button:"))
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
