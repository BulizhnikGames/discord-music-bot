package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func ShuffleInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.ShuffleQueue(interaction.GuildID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, "🔀  shuffled  🔀")
		return nil
	case discordgo.InteractionMessageComponent:
		err := bot.ShuffleQueue(interaction.GuildID)
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, "🔀  shuffled  🔀")
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
