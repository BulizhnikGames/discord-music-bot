package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/go-faster/errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Downloaded *internal.Song
	Err        error
}

func Download(ctx context.Context, guildID, query string, res chan<- Result) {
	start := time.Now()
	firstArg := query
	if !strings.HasPrefix(query, config.LINK_PREFIX) {
		firstArg = fmt.Sprintf("ytsearch10:%s", strings.ReplaceAll(query, "\"", ""))
	}
	args := []string{
		firstArg,
		//"-N", "4",
		"--extract-audio",
		"--buffer-size", "4096",
		"--retries", "1",
		"--audio-format", "opus",
		"--no-playlist",
		"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
		"--max-downloads", "1",
		"--output", fmt.Sprintf("%s%s/%d-%%(id)s.opus", config.Storage, guildID, start.Unix()),
		"--quiet",
		"--print-json",
		"--ignore-errors", // Ignores unavailable videos
		"--no-color",
		"--no-check-formats",
	}
	log.Printf("yt-dlp %s", strings.Join(args, " "))
	var commandPath string
	if config.Utils == "" {
		commandPath = "yt-dlp"
	} else {
		commandPath = config.Utils + "yt-dlp.exe"
	}
	cmd := exec.Command(commandPath, args...)
	if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
		res <- Result{nil, errors.Errorf("failed to search and download audio (query %s): %v", query, err)}
	} else {
		videoMetadata := internal.VideoMetadata{}
		err = json.Unmarshal(data, &videoMetadata)
		if err != nil {
			res <- Result{nil, errors.Errorf("failed to unmarshal video metadata (query %s): %v", query, err)}
		}
		//dotIdx := strings.LastIndex(videoMetadata.Filename, ".")
		//slashIdx := strings.LastIndex(videoMetadata.Filename, `\`)
		path := fmt.Sprintf("%s%s/%d-%s.opus", config.Storage, guildID, start.Unix(), videoMetadata.ID)
		select {
		case <-ctx.Done():
			os.RemoveAll(path)
			res <- Result{nil, ctx.Err()}
		default:
			res <- Result{&internal.Song{
				Title:    videoMetadata.Title,
				Author:   videoMetadata.Uploader,
				URL:      videoMetadata.URL,
				Query:    query,
				FilePath: path,
				Duration: videoMetadata.Duration,
			}, nil}
		}
	}
}
