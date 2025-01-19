package main

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions/middleware"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"github.com/redis/go-redis/v9"
	"os"
	"os/signal"
)

// TODO: add /help

// TODO: download age restricted content

// TODO: think of speeding up search for autocompletion

func main() {
	cfg := config.LoadConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Url,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DBid,
	})

	discordBot := bot.Init(cfg, redisClient, interactions.InitialResponse)

	djMustInChannel := func(next servers.InteractionFunc) servers.InteractionFunc {
		return middleware.DJOrAdminOnly(middleware.ActiveChannelOnly(next, true))
	}

	discordBot.RegisterCommand(
		"play",
		middleware.DJOrAdminOnly(
			middleware.ActiveChannelOnly(
				interactions.PlayInteraction(youtube.Search),
				false,
			),
		),
	)
	discordBot.RegisterCommand("leave", djMustInChannel(interactions.LeaveInteraction))
	discordBot.RegisterCommand("clear", djMustInChannel(interactions.ClearInteraction))
	discordBot.RegisterCommand("stop", djMustInChannel(interactions.ClearInteraction))
	discordBot.RegisterCommand("skip", djMustInChannel(interactions.SkipInteraction))
	discordBot.RegisterCommand("shuffle", djMustInChannel(interactions.ShuffleInteraction))
	discordBot.RegisterCommand("queue", middleware.ActiveChannelOnly(interactions.QueueInteraction, false))
	discordBot.RegisterCommand("nowplaying", middleware.ActiveChannelOnly(interactions.NowPlayingInteraction, false))
	discordBot.RegisterCommand("loop", djMustInChannel(interactions.LoopInteraction))
	discordBot.RegisterCommand("pause", djMustInChannel(interactions.PauseInteraction))
	discordBot.RegisterCommand("resume", djMustInChannel(interactions.ResumeInteraction))
	discordBot.RegisterCommand("dj-mode", middleware.AdminOnly(interactions.DJModeInteraction))
	discordBot.RegisterCommand("dj-off", middleware.AdminOnly(interactions.NoDJInteraction))

	discordBot.RegisterCommand("loop0", djMustInChannel(interactions.Loop0))
	discordBot.RegisterCommand("loop1", djMustInChannel(interactions.Loop1))
	discordBot.RegisterCommand("loop2", djMustInChannel(interactions.Loop2))

	discordBot.RegisterCommand("queueprev", middleware.ActiveChannelOnly(interactions.QueuePrevInteraction, false))
	discordBot.RegisterCommand("queuenext", middleware.ActiveChannelOnly(interactions.QueueNextInteraction, false))

	go discordBot.Run()

	fmt.Println("Bot is now running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	discordBot.Stop()
}
