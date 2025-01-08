package youtube

import (
	"encoding/json"
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/go-faster/errors"
	"log"
	"os/exec"
	"strings"
	"time"
)

func Download(query string) (*internal.Song, error) {
	start := time.Now()
	youtubeDownloader, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, errors.New("yt-dlp not found in path")
	} else {
		firstArg := query
		if !strings.HasPrefix(query, config.LINK_PREFIX) {
			firstArg = fmt.Sprintf("ytsearch10:%s", strings.ReplaceAll(query, "\"", ""))
		}
		args := []string{
			firstArg,
			"--extract-audio",
			"--audio-format", "opus",
			"--no-playlist",
			"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
			"--max-downloads", "1",
			"--output", fmt.Sprintf("./storage/%d-%%(id)s.opus", start.Unix()),
			"--quiet",
			"--print-json",
			"--ignore-errors", // Ignores unavailable videos
			"--no-color",
			"--no-check-formats",
		}
		log.Printf("yt-dlp %s", strings.Join(args, " "))
		cmd := exec.Command(youtubeDownloader, args...)
		if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
			return nil, errors.Errorf("failed to search and download audio: %v", err)
		} else {
			videoMetadata := internal.VideoMetadata{}
			err = json.Unmarshal(data, &videoMetadata)
			if err != nil {
				return nil, errors.Errorf("failed to unmarshal video metadata: %v", err)
			}
			dotIdx := strings.LastIndex(videoMetadata.Filename, ".")
			slashIdx := strings.LastIndex(videoMetadata.Filename, `\`)
			return &internal.Song{
				Title:    videoMetadata.Title,
				Author:   videoMetadata.Uploader,
				URL:      videoMetadata.URL,
				FilePath: config.Storage + videoMetadata.Filename[slashIdx+1:dotIdx] + `.opus`,
				Duration: videoMetadata.Duration,
			}, nil
		}
	}
}
