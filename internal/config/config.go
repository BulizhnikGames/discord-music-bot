package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const (
	QUEUE_SIZE  = 140
	LINK_PREFIX = "https://www.youtube.com/watch?v="
)

var (
	Tools string
	Logs  string
)

//var Cookies string
//var CookiesGuildID string

type RedisConfig struct {
	Url      string
	DBid     int
	Username string
	Password string
}

type Config struct {
	BotToken    string
	AppID       string
	SearchLimit int
	Redis       RedisConfig
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

	Tools = os.Getenv("TOOLS_PATH")
	if len(Tools) > 0 && Tools[len(Tools)-1] != '/' {
		Tools = Tools + "/"
	}

	Logs = os.Getenv("LOGS_PATH")
	if len(Logs) > 0 && Logs[len(Logs)-1] != '/' {
		Logs = Logs + "/"
	}

	log.Printf("Logging path: <%s>", Logs)

	cfg.Redis.Url = os.Getenv("DB_URL")
	if cfg.Redis.Url == "" {
		log.Fatal("Redis url not found in .env")
	}

	dbIDStr := os.Getenv("DB_ID")
	if dbIDStr == "" {
		log.Fatal("Redis db id not found in .env")
	}
	if cfg.Redis.DBid, err = strconv.Atoi(dbIDStr); err != nil {
		log.Fatalf("Error parsing redis db id to int: %v", err)
	}

	cfg.Redis.Username = os.Getenv("DB_USERNAME")
	cfg.Redis.Password = os.Getenv("DB_PASSWORD")

	//Cookies = os.Getenv("COOKIES")
	//CookiesGuildID = os.Getenv("COOKIES_GUILD_ID")

	return cfg
}
