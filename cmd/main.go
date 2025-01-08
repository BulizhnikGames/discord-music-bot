package main

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions"
	"os"
	"os/signal"
)

func main() {
	cfg := config.LoadConfig()

	discordBot := bot.Init(cfg.BotToken, cfg.AppID)

	discordBot.RegisterCommand("play", interactions.PlayInteraction)
	discordBot.RegisterCommand("leave", interactions.LeaveInteraction)
	discordBot.RegisterCommand("clear", interactions.ClearInteraction)

	go discordBot.Run()

	fmt.Println("Bot is now running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	discordBot.Stop()
}
