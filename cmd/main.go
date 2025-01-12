package main

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions"
	"github.com/BulizhnikGames/discord-music-bot/internal/interactions/middleware"
	"os"
	"os/signal"
)

// TODO: messages formating

// TODO: dj role
// TODO: select text channel

// TODO: add /help or /info

func main() {
	cfg := config.LoadConfig()

	discordBot := bot.Init(cfg, interactions.InitialResponse)

	discordBot.RegisterCommand("play", middleware.ActiveChannelOnly(interactions.PlayInteraction, false))
	discordBot.RegisterCommand("leave", middleware.ActiveChannelOnly(interactions.LeaveInteraction, true))
	discordBot.RegisterCommand("clear", middleware.ActiveChannelOnly(interactions.ClearInteraction, true))
	discordBot.RegisterCommand("stop", middleware.ActiveChannelOnly(interactions.ClearInteraction, true))
	discordBot.RegisterCommand("skip", middleware.ActiveChannelOnly(interactions.SkipInteraction, true))
	discordBot.RegisterCommand("shuffle", middleware.ActiveChannelOnly(interactions.ShuffleInteraction, true))
	discordBot.RegisterCommand("queue", middleware.ActiveChannelOnly(interactions.QueueInteraction, false))
	discordBot.RegisterCommand("nowplaying", middleware.ActiveChannelOnly(interactions.NowPlayingInteraction, false))
	discordBot.RegisterCommand("loop", middleware.ActiveChannelOnly(interactions.LoopInteraction, true))
	discordBot.RegisterCommand("pause", middleware.ActiveChannelOnly(interactions.PauseInteraction, true))
	discordBot.RegisterCommand("resume", middleware.ActiveChannelOnly(interactions.ResumeInteraction, true))

	go discordBot.Run()

	fmt.Println("Bot is now running.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	discordBot.Stop()
}
