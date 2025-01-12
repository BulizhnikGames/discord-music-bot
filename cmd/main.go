package main

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions/middleware"
	"github.com/redis/go-redis/v9"
	"os"
	"os/signal"
)

// TODO: messages formating

// TODO: select text channel

// TODO: send dj role normally

// TODO: add /help or /info

func main() {
	cfg := config.LoadConfig()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Url,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DBid,
	})

	discordBot := bot.Init(cfg, redisClient, interactions.InitialResponse)

	djMustInChannel := func(next bot.InteractionFunc) bot.InteractionFunc {
		return middleware.DJOrAdminOnly(middleware.ActiveChannelOnly(next, true))
	}

	discordBot.RegisterCommand(
		"play",
		middleware.DJOrAdminOnly(
			middleware.ActiveChannelOnly(
				interactions.PlayInteraction,
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

	go discordBot.Run()

	fmt.Println("Bot is now running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	discordBot.Stop()
}
