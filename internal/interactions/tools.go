package interactions

import (
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/bwmarrin/discordgo"
	"log"
)

func responseToInteraction(bot *bot.DiscordBot, interaction *discordgo.InteractionCreate, text string, flags discordgo.MessageFlags) {
	err := bot.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Flags:   flags,
		},
	})
	if err != nil {
		log.Printf("Failed to respond to interaction: %v", err)
	}
}
