package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube/api"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"log"
	"sync"
)

type InteractionFunc func(bot *DiscordBot, interaction *discordgo.InteractionCreate) error

type VoiceEntity struct {
	VoiceConnection *discordgo.VoiceConnection
	Queue           chan string
	Ctx             context.Context
	Stop            context.CancelFunc
}

type DiscordBot struct {
	Session       *discordgo.Session
	Youtube       *api.Youtube
	Interactions  map[string]InteractionFunc
	VoiceEntities internal.AsyncMap[string, *VoiceEntity]
}

func Init(BotToken, AppID string) *DiscordBot {
	session, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %s", err)
	}

	_, err = session.ApplicationCommandBulkOverwrite(AppID, "", []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "play YT video by name or URL",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "name",
					Description:  "name or url of the video",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "leave",
			Description: "leave voice chat",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "clear",
			Description: "clear playback queue",
			Type:        discordgo.ChatApplicationCommand,
		},
	})
	if err != nil {
		log.Fatalf("Error initializing application's slash interactions: %s", err)
	}

	bot := &DiscordBot{
		Session:      session,
		Youtube:      api.NewService(),
		Interactions: make(map[string]InteractionFunc),
		VoiceEntities: internal.AsyncMap[string, *VoiceEntity]{
			Data:  make(map[string]*VoiceEntity, 200),
			Mutex: &sync.RWMutex{},
		},
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := bot.Interactions[i.ApplicationCommandData().Name]; ok {
			go func() {
				err := handler(bot, i)
				if err != nil {
					var logErr, userErr error
					if errors.Unwrap(err) == nil {
						logErr = err
						userErr = errors.New("Internal error")
					} else {
						userErr = err
						logErr = errors.Unwrap(err)
					}
					log.Printf("Error executing interaction (%s %s): %v", i.ApplicationCommandData().Name, i.Type.String(), logErr)
					_ = session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: userErr.Error(),
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				}
			}()
		} else {
			log.Printf("Error: no such command: %s", i.ApplicationCommandData().Name)
		}
	})

	return bot
}

func (bot *DiscordBot) RegisterCommand(name string, handler InteractionFunc) {
	bot.Interactions[name] = handler
}

func (bot *DiscordBot) Run() {
	err := bot.Session.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %s", err)
	}
	defer bot.Stop()
}

func (bot *DiscordBot) Stop() {
	bot.VoiceEntities.Mutex.Lock()
	for _, voice := range bot.VoiceEntities.Data {
		err := voice.VoiceConnection.Disconnect()
		if err != nil {
			log.Printf("Couldn't disconnect from voice chat (id: %s, guild: %s)",
				voice.VoiceConnection.ChannelID,
				voice.VoiceConnection.GuildID,
			)
		}
	}
	bot.VoiceEntities.Mutex.Unlock()
	bot.Session.Close()
}
