package servers

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers/voice"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strings"
)

type InteractionFunc func(server *Server, interaction *discordgo.InteractionCreate) error
type InteractionMiddleware func(next InteractionFunc, arg any) InteractionFunc

type ResponseFunc func(*Server, *discordgo.InteractionCreate)

type Server struct {
	VoiceChat    *voice.Connection
	Interactions chan *discordgo.InteractionCreate
	Session      *discordgo.Session
	interactions map[string]InteractionFunc
	GuildID      string
	db           *redis.Client
	Logger       *log.Logger
}

func New(s *discordgo.Session, iMap map[string]InteractionFunc, id string, db *redis.Client, r ResponseFunc) *Server {
	var logger *log.Logger
	if config.Logs != "" {
		file, err := os.Create(config.Logs + id + ".txt")
		if err != nil {
			log.Printf("Error creating log file: %v", err)
			logger = log.New(os.Stdout, "", log.LstdFlags)
		} else {
			logger = log.New(file, "", log.LstdFlags)
		}
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	server := &Server{
		Interactions: make(chan *discordgo.InteractionCreate, 10),
		Session:      s,
		interactions: iMap,
		GuildID:      id,
		db:           db,
		Logger:       logger,
	}
	go server.Run(r)
	return server
}

func (server *Server) Run(initResp ResponseFunc) {
	defer func() {
		err := recover()
		if err != nil {
			server.Logger.Printf("panic recovered: %v", err)
			server.Run(initResp)
		}
	}()

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
			go initResp(server, interaction)
			if interaction.Type == discordgo.InteractionApplicationCommandAutocomplete {
				go server.Handle(handler, interaction, name)
			} else {
				server.Handle(handler, interaction, name)
			}
		} else {
			server.Logger.Printf("Error: no such command: %s", name)
		}
	}
}

func (server *Server) Handle(handler InteractionFunc, interaction *discordgo.InteractionCreate, name string) {
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
		server.Logger.Printf(
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
}
