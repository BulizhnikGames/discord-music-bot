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
	"runtime"
	"strings"
)

type Result struct {
	Downloaded *internal.Song
	Err        error
}

func Download(ctx context.Context, query string, res chan<- Result) {
	firstArg := query
	if !strings.HasPrefix(query, config.LINK_PREFIX) {
		firstArg = fmt.Sprintf("ytsearch10:%s", strings.ReplaceAll(query, "\"", ""))
	}
	args := []string{
		firstArg,
		//"-N", "4",
		"-f", "bestaudio",
		"--max-downloads", "1",
		"--no-playlist",
		"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
		"--skip-download",
		"--print-json",
	}
	log.Printf("yt-dlp %s", strings.Join(args, " "))
	var commandPath string
	if config.Utils == "" {
		commandPath = "yt-dlp"
	} else {
		commandPath = config.Utils + "yt-dlp.exe"
	}
	cmd := exec.Command(commandPath, args...)
	if runtime.GOOS == "windows" {
		cmd.Env = append(os.Environ(), "LANG=ru_RU.UTF-8")
	}
	if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
		res <- Result{nil, errors.Errorf("failed to search and download audio (query %s): %v", query, err)}
		return
	} else {
		videoMetadata := internal.VideoMetadata{}
		err = json.Unmarshal(data, &videoMetadata)
		if err != nil {
			res <- Result{nil, errors.Errorf("failed to unmarshal video metadata (query %s): %v", query, err)}
			return
		}
		select {
		case <-ctx.Done():
			res <- Result{nil, ctx.Err()}
			return
		default:
			res <- Result{&internal.Song{
				Title:        videoMetadata.Title,
				Author:       videoMetadata.Uploader,
				Duration:     videoMetadata.Duration,
				FileURL:      videoMetadata.URL,
				ThumbnailUrl: videoMetadata.Thumbnail,
				Query:        query,
			}, nil}
			return
		}
	}
}
