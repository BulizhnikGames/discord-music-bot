package interactions

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
)

func LoopInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) error {
	switch interaction.Type {
	case discordgo.InteractionApplicationCommand:
		loop := interaction.ApplicationCommandData().Options[0].IntValue()
		err := bot.SetLoop(interaction.GuildID, int(loop))
		if err != nil {
			return err
		}
		switch loop {
		case 0:
			responseToInteraction(bot, interaction, fmt.Sprintf(":ballot_box_with_check: looping disabled :ballot_box_with_check:"), 0)
		case 1:
			responseToInteraction(bot, interaction, fmt.Sprintf(":repeat: looping over queue :repeat:"), 0)
		case 2:
			responseToInteraction(bot, interaction, fmt.Sprintf(":repeat_one: looping over song :repeat_one:"), 0)
		default:
			responseToInteraction(bot, interaction, fmt.Sprintf(":ballot_box_with_check: looping disabled :ballot_box_with_check:"), 0)
		}
		return nil
	default:
		return errors.Errorf("unknown interaction type: %s", interaction.Type.String())
	}
}
