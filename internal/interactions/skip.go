package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func SkipInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.SkipSong(interaction.GuildID, "")
		if err != nil {
			return err
		}
		responseToInteraction(bot, interaction, "⏩ skipped")
		return nil
	case discordgo.InteractionMessageComponent:
		return bot.SkipSong(interaction.GuildID, "⏩ skipped")
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
