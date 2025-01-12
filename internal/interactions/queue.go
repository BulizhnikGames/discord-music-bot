package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

func QueueInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		queue, err := bot.GetQueue(interaction.GuildID)
		if err != nil {
			return err
		}
		if len(queue) == 0 {
			responseToInteraction(bot, interaction, ":musical_note:  playback queue is empty  :musical_note:")
			return nil
		}
		maxLength := 0
		for i, song := range queue {
			maxLength = max(maxLength, utf8.RuneCountInString(song)+utf8.RuneCountInString(strconv.Itoa(i))+2)
		}
		message := strings.Builder{}
		message.WriteString(":musical_note:  playback queue  :musical_note:\n")
		for i, song := range queue {
			message.WriteRune('`')
			message.WriteString(fmt.Sprintf("%d. %s", i+1, song))
			diff := maxLength - (utf8.RuneCountInString(song) + utf8.RuneCountInString(strconv.Itoa(i)) + 2)
			for j := 0; j < diff; j++ {
				message.WriteRune(' ')
			}
			message.WriteRune('`')
			if i != len(queue)-1 {
				message.WriteString("\n")
			}
		}
		responseToInteraction(bot, interaction, message.String())
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
				":musical_note:  now playing `%s | %d:%02d / %d:%02d` by `%s`  :musical_note:",
				song.Title,
				curr/60,
				curr%60,
				song.Duration/60,
				song.Duration%60,
				song.Author,
			)
			responseToInteraction(bot, interaction, message)
		} else {
			responseToInteraction(bot, interaction, ":mute: nothing is playing :mute:")
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
