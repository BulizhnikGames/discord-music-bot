package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func QueueInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		queue, err := bot.GetQueue(interaction.GuildID)
		if err != nil {
			return err
		}
		if len(queue) == 0 {
			responseToInteraction(bot, interaction, "🎵  playback queue is empty  🎵")
			return nil
		}
		resp := make([]string, 1, len(queue)+1)
		resp[0] = "🎵  playback queue  🎵"
		for i, song := range queue {
			resp = append(resp, fmt.Sprintf("%d. %s", i+1, song))
		}
		responseToInteractionWithList(bot, interaction, resp...)
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}

func NowPlayingInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		song, curr, err := bot.NowPlaying(interaction.GuildID)
		if err != nil {
			return err
		}
		if song != nil {
			message := fmt.Sprintf(
				"🎵  now playing `%s | %d:%02d / %d:%02d` by `%s`  🎵",
				song.Title,
				curr/60,
				curr%60,
				song.Duration/60,
				song.Duration%60,
				song.Author,
			)
			responseToInteraction(bot, interaction, message)
		} else {
			responseToInteraction(bot, interaction, "🔇  nothing is playing  🔇")
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
