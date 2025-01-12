package bot

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube/api"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"log"
	"os"
	"sync"
)

type InteractionFunc func(bot *DiscordBot, interaction *discordgo.InteractionCreate) error
type InteractionMiddleware func(next InteractionFunc, arg any) InteractionFunc

type VoiceEntity struct {
	voiceConnection *discordgo.VoiceConnection
	queue           *internal.MusicQueue
	nowPlaying      *internal.PlayingSong
	cache           internal.AsyncMap[string, *internal.SongCache] // key is user's query for song
	loop            int                                            // 0 - no loop, 1 - queue loop, 2 - single loop
	textChannel     string
	stop            context.CancelFunc
}

type DiscordBot struct {
	Session       *discordgo.Session
	Youtube       *api.Youtube
	Interactions  map[string]InteractionFunc
	VoiceEntities internal.AsyncMap[string, *VoiceEntity]
}

func Init(cfg config.Config, initialResp func(*DiscordBot, *discordgo.InteractionCreate)) *DiscordBot {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %s", err)
	}

	_, err = session.ApplicationCommandBulkOverwrite(cfg.AppID, "", []*discordgo.ApplicationCommand{
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
		{
			Name:        "stop",
			Description: "stop playback and clear queue",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "skip",
			Description: "skip current song",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "shuffle",
			Description: "shuffle queue",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "queue",
			Description: "get songs in queue",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "nowplaying",
			Description: "get current song",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "loop",
			Description: "0 - no loop, 1 - loop queue, 2 - loop current song",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "arg",
					Description:  "loop command argument",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "pause",
			Description: "pause playback",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "resume",
			Description: "resume playback",
			Type:        discordgo.ChatApplicationCommand,
		},
	})
	if err != nil {
		log.Fatalf("Error initializing application's slash interactions: %s", err)
	}

	bot := &DiscordBot{
		Session:      session,
		Youtube:      api.NewService(cfg.SearchLimit),
		Interactions: make(map[string]InteractionFunc),
		VoiceEntities: internal.AsyncMap[string, *VoiceEntity]{
			Data:  make(map[string]*VoiceEntity),
			Mutex: &sync.RWMutex{},
		},
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := bot.Interactions[i.ApplicationCommandData().Name]; ok {
			go func() {
				go initialResp(bot, i)
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
		err := voice.voiceConnection.Disconnect()
		if err != nil {
			log.Printf("Couldn't disconnect from voice chat (id: %s, guild: %s)",
				voice.voiceConnection.ChannelID,
				voice.voiceConnection.GuildID,
			)
		}
	}
	bot.VoiceEntities.Mutex.Unlock()
	bot.Session.Close()
	os.RemoveAll(config.Storage)
	//os.Mkdir(config.Storage, 0777)
}
