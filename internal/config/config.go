package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const QUEUE_SIZE = 140
const LINK_PREFIX = "https://www.youtube.com/watch?v="

var Storage string
var Utils string

type Config struct {
	BotToken    string
	AppID       string
	SearchLimit int
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

	cfg.SearchLimit, err = strconv.Atoi(os.Getenv("SEARCH_LIMIT"))
	if err != nil {
		log.Fatalf("Incorrect search limit: %s", err)
	}

	Storage = os.Getenv("STORAGE")
	if Storage == "" {
		log.Fatal("Storage not found")
	}
	if Storage[len(Storage)-1] != '/' {
		Storage = Storage + "/"
	}

	Utils = os.Getenv("UTILS_PATH")
	if Utils == "" {
		log.Fatal("Utils not found")
	}
	if Utils[len(Utils)-1] != '/' {
		Utils = Utils + "/"
	}

	return cfg
}
