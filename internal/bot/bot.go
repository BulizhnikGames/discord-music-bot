package bot

import (
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
)

type DiscordBot struct {
	Session      *discordgo.Session
	interactions map[string]servers.InteractionFunc
	servers      internal.AsyncMap[string, *servers.Server]
}

func Init(cfg config.Config, db *redis.Client, respFunc servers.ResponseFunc) *DiscordBot {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %s", err)
	}

	_, err = session.ApplicationCommandBulkOverwrite(cfg.AppID, "", []*discordgo.ApplicationCommand{
		{Name: "clear", Description: "clear playback queue", Type: discordgo.ChatApplicationCommand},
		{
			Name:        "dj-mode",
			Description: "set dj mode, only members with dj role can manipulate queue and playback",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "role",
					Description:  "dj role",
					Type:         discordgo.ApplicationCommandOptionRole,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{Name: "dj-off", Description: "turn off dj-mode", Type: discordgo.ChatApplicationCommand},
		{Name: "help", Description: "get list of commands and their meanings", Type: discordgo.ChatApplicationCommand},
		{Name: "leave", Description: "leave voice chat", Type: discordgo.ChatApplicationCommand},
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
		{Name: "nowplaying", Description: "get current song", Type: discordgo.ChatApplicationCommand},
		{Name: "pause", Description: "pause playback", Type: discordgo.ChatApplicationCommand},
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
		{Name: "queue", Description: "get songs in queue", Type: discordgo.ChatApplicationCommand},
		{Name: "resume", Description: "resume playback", Type: discordgo.ChatApplicationCommand},
		{Name: "shuffle", Description: "shuffle queue", Type: discordgo.ChatApplicationCommand},
		{Name: "skip", Description: "skip current song", Type: discordgo.ChatApplicationCommand},
		{Name: "stop", Description: "clear playback queue", Type: discordgo.ChatApplicationCommand},
	})
	if err != nil {
		log.Fatalf("Error initializing application's slash interactions: %s", err)
	}

	bot := &DiscordBot{
		Session:      session,
		interactions: make(map[string]servers.InteractionFunc),
		servers: internal.AsyncMap[string, *servers.Server]{
			Data:  make(map[string]*servers.Server),
			Mutex: &sync.RWMutex{},
		},
	}

	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		serverID := interaction.GuildID
		bot.servers.Mutex.RLock()
		if serv, ok := bot.servers.Data[serverID]; ok {
			go func() {
				serv.Interactions <- interaction
			}()
			bot.servers.Mutex.RUnlock()
		} else {
			bot.servers.Mutex.RUnlock()
			bot.servers.Mutex.Lock()
			bot.servers.Data[serverID] = servers.New(session, bot.interactions, serverID, db, respFunc)
			interactionChan := bot.servers.Data[serverID].Interactions
			bot.servers.Mutex.Unlock()
			go func() {
				interactionChan <- interaction
			}()
		}
	})

	session.AddHandler(func(session *discordgo.Session, update *discordgo.VoiceStateUpdate) {
		if update.UserID != session.State.User.ID {
			return
		}
		if update.ChannelID == "" { // Disconnected
			bot.servers.Mutex.RLock()
			defer bot.servers.Mutex.RUnlock()
			if serv, ok := bot.servers.Data[update.GuildID]; ok {
				_ = serv.TryLeaveVoiceChat()
			}
		}
	})

	return bot
}

func (bot *DiscordBot) RegisterCommand(name string, handler servers.InteractionFunc) {
	bot.interactions[name] = handler
}

func (bot *DiscordBot) Run() {
	err := bot.Session.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %s", err)
	}
	defer bot.Stop()
}

func (bot *DiscordBot) Stop() {
	bot.servers.Mutex.Lock()
	for _, server := range bot.servers.Data {
		err := server.TryLeaveVoiceChat()
		if err != nil {
			log.Printf("Error leaving voice chat (guild: %s): %v", server.GuildID, err)
		}
	}
	bot.servers.Mutex.Unlock()
	bot.Session.Close()
}
