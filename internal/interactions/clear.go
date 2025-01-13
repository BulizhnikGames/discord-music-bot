package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func ClearInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.ClearQueue(interaction.GuildID, "")
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, "⏹️  queue cleared  ⏹️")
		return nil
	case discordgo.InteractionMessageComponent:
		return bot.ClearQueue(interaction.GuildID, "⏹️  queue cleared  ⏹️")
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
