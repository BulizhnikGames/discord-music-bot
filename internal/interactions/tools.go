package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"log"
)

func InitialResponse(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate) {
	if interaction.Type == discordgo.InteractionApplicationCommand {
		err := bot.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: ":hourglass: processing... :hourglass_flowing_sand:",
			},
		})
		if err != nil {
			log.Printf("Failed to initialy respond to interaction: %v", err)
		}
	}
}

func responseToInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate, text string) {
	_, err := bot.Session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &text,
	})
	if err != nil {
		log.Printf("Failed to respond to interaction (with text %s): %v", text, err)
	}
}
