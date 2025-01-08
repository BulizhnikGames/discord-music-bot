package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	BotToken string
	AppID    string
}

func LoadConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	var cfg Config

	cfg.BotToken = os.Getenv("BOT_TOKEN")
	if cfg.BotToken == "" {
		log.Fatal("Bot token not found")
	}

	cfg.AppID = os.Getenv("APP_ID")
	if cfg.AppID == "" {
		log.Fatal("App ID not found")
	}

	return cfg
}
