package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func NowPlayingInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		song, err := bot.NowPlaying(interaction.GuildID)
		if err != nil {
			return err
		}
		if song != nil {
			message := fmt.Sprintf(
				":musical_note: now playing `%s | %d:%02d` by `%s` :musical_note:",
				song.Title,
				song.Duration/60,
				song.Duration%60,
				song.Author,
			)
			responseToInteraction(bot, interaction, message, 0)
		} else {
			responseToInteraction(bot, interaction, ":mute: nothing is playing :mute:", 0)
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
