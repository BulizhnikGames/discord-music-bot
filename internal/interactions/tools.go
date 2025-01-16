package interactions

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func InitialResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type == discordgo.InteractionApplicationCommand {
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			//Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, // may be use this method
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Author: &discordgo.MessageEmbedAuthor{
							Name:    "⏳  processing...  ⌛",
							IconURL: interaction.Member.User.AvatarURL("64x64"),
						},
						Color: 2326507,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "github.com/BulizhnikGames/discord-music-bot",
						},
					},
				},
			},
		})
		if err != nil {
			log.Printf("Failed to initialy respond to interaction: %v", err)
		}
	} else if interaction.Type == discordgo.InteractionMessageComponent {
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		if err != nil {
			log.Printf("Failed to initialy respond to interaction: %v", err)
		}
	}
}

func responseToInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate, elems ...string) {
	for len(elems) < 5 {
		elems = append(elems, "")
	}
	fields := make([]*discordgo.MessageEmbedField, 0, 1)
	if elems[3] != "" {
		fields = append(fields, &discordgo.MessageEmbedField{Name: elems[3]})
	}

	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Author: &discordgo.MessageEmbedAuthor{
					Name:    elems[0],
					IconURL: interaction.Member.User.AvatarURL("64x64"),
				},
				Title:  elems[1],
				URL:    elems[2],
				Color:  2326507,
				Fields: fields,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: elems[4],
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "github.com/BulizhnikGames/discord-music-bot",
				},
			},
		},
		//Components: &[]discordgo.MessageComponent{},
	})
	if err != nil {
		log.Printf("Failed to respond to interaction (with text %s): %v", elems[0], err)
	}
}
