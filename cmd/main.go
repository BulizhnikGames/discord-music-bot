package main

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions"
	"os"
	"os/signal"
)

// TODO: errors
// TODO: messages formating
// TODO: middleware
// TODO: now playing
// TODO: playback optimization

func main() {
	cfg := config.LoadConfig()

	discordBot := bot.Init(cfg.BotToken, cfg.AppID, cfg.SearchLimit)

	discordBot.RegisterCommand("play", interactions.PlayInteraction)
	discordBot.RegisterCommand("leave", interactions.LeaveInteraction)
	discordBot.RegisterCommand("clear", interactions.ClearInteraction)
	discordBot.RegisterCommand("stop", interactions.StopInteraction)
	discordBot.RegisterCommand("skip", interactions.SkipInteraction)
	discordBot.RegisterCommand("shuffle", interactions.ShuffleInteraction)
	discordBot.RegisterCommand("queue", interactions.QueueInteraction)

	go discordBot.Run()

	fmt.Println("Bot is now running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	discordBot.Stop()
}
