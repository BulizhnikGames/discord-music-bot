package servers

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers/voice"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"log"
	"strings"
)

type InteractionFunc func(server *Server, interaction *discordgo.InteractionCreate) error
type InteractionMiddleware func(next InteractionFunc, arg any) InteractionFunc

type ResponseFunc func(*discordgo.Session, *discordgo.InteractionCreate)

type Server struct {
	VoiceChat    *voice.Connection
	Interactions chan *discordgo.InteractionCreate
	Session      *discordgo.Session
	interactions map[string]InteractionFunc
	GuildID      string
	db           *redis.Client
}

func New(s *discordgo.Session, iMap map[string]InteractionFunc, id string, db *redis.Client, r ResponseFunc) *Server {
	server := &Server{
		Interactions: make(chan *discordgo.InteractionCreate, 10),
		Session:      s,
		interactions: iMap,
		GuildID:      id,
		db:           db,
	}
	go server.Run(r)

	return server
}

func (server *Server) Run(initResp ResponseFunc) {
	for interaction := range server.Interactions {
		var name string
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			name = interaction.ApplicationCommandData().Name
		case discordgo.InteractionApplicationCommandAutocomplete:
			name = interaction.ApplicationCommandData().Name
		case discordgo.InteractionMessageComponent:
			name = interaction.MessageComponentData().CustomID
		default:
			name = ""
		}
		if idx := strings.Index(name, ":"); idx != -1 {
			if name[:idx] == server.Session.State.SessionID {
				name = name[idx+1:]
			} else {
				name = ""
			}
		}
		if handler, ok := server.interactions[name]; ok {
			go func() {
				go initResp(server.Session, interaction)
				err := handler(server, interaction)
				if err != nil {
					var userErr, logErr error
					if det, ok := err.(*errors.DetailedError); ok {
						userErr = det.User
						logErr = det.Log
					} else {
						logErr = err
						userErr = errors.New("internal error")
					}
					log.Printf(
						"Error executing interaction (%s %s): %s",
						name,
						interaction.Type.String(),
						logErr.Error(),
					)
					if interaction.Type == discordgo.InteractionApplicationCommand {
						resp := fmt.Sprintf("❌  %s  ❌", userErr.Error())
						_, _ = server.Session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
							Embeds: &[]*discordgo.MessageEmbed{
								{
									Author: &discordgo.MessageEmbedAuthor{
										Name:    resp,
										IconURL: interaction.Member.User.AvatarURL("64x64"),
									},
									Color: 2326507,
									Footer: &discordgo.MessageEmbedFooter{
										Text: "github.com/BulizhnikGames/discord-music-bot",
									},
								},
							},
						})
					}
				}
			}()
		} else {
			log.Printf("Error: no such command: %s", name)
		}
	}
}

func (server *Server) Stop() {
	if server.VoiceChat != nil {
		server.VoiceChat.Queue.Clear()
		server.VoiceChat.Mutex.Lock()
		if server.VoiceChat.NowPlaying != nil {
			server.VoiceChat.NowPlaying.EncodeSession.Cleanup()
		}
		server.VoiceChat.Mutex.Unlock()
		err := server.VoiceChat.VoiceConnection.Disconnect()
		if err != nil {
			log.Printf("Couldn't disconnect from voice chat (id: %s, guild: %s)",
				server.VoiceChat.VoiceConnection.ChannelID,
				server.VoiceChat.VoiceConnection.GuildID,
			)
		}
	}
}
