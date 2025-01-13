package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func PauseInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.Pause(interaction.GuildID, true)
		if err != nil {
			return err
		}
		go bot.TryRegenPlaybackMessage(interaction.GuildID)
		responseToInteraction(bot, interaction, "⏸️  playback paused  ⏸️")
		return nil
	case discordgo.InteractionMessageComponent:
		err := bot.Pause(interaction.GuildID, true)
		if err != nil {
			return err
		}
		go bot.TryRegenPlaybackMessage(interaction.GuildID)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func ResumeInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		err := bot.Pause(interaction.GuildID, false)
		if err != nil {
			return err
		}
		go bot.TryRegenPlaybackMessage(interaction.GuildID)
		responseToInteraction(bot, interaction, "▶️  playback resumed  ▶️")
		return nil
	case discordgo.InteractionMessageComponent:
		err := bot.Pause(interaction.GuildID, false)
		if err != nil {
			return err
		}
		go bot.TryRegenPlaybackMessage(interaction.GuildID)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
