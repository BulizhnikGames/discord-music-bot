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
	Host     string
	Port     string
	Username string
	Password string
	DB       int
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
	if len(Logs) > 0 {
		if Logs[len(Logs)-1] != '/' {
			Logs = Logs + "/"
		}
		if err = os.MkdirAll(Logs, 0755); err == nil {
			log.Printf("Logging path: <%s>", Logs)
		} else {
			log.Fatalf("Couldn't create logging dir: %v", err)
		}
	}

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	if cfg.Redis.Host == "" {
		log.Fatal("Redis host not found")
	}

	cfg.Redis.Port = os.Getenv("REDIS_PORT")
	if cfg.Redis.Port == "" {
		log.Fatal("Redis port not found")
	}

	dbIDStr := os.Getenv("REDIS_DB_ID")
	if dbIDStr == "" {
		dbIDStr = "0"
	}
	if cfg.Redis.DB, err = strconv.Atoi(dbIDStr); err != nil {
		log.Fatalf("Error parsing redis db id to int: %v", err)
	}

	cfg.Redis.Username = os.Getenv("REDIS_USERNAME")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	//Cookies = os.Getenv("COOKIES")
	//CookiesGuildID = os.Getenv("COOKIES_GUILD_ID")

	return cfg
}
