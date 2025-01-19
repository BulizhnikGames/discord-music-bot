package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/go-faster/errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type MetadataResult struct {
	Data *internal.Song
	Err  error
}

func GetMetadataWithContext(ctx context.Context, query string, res chan<- MetadataResult) {
	song, err := GetMetadata(query)
	if ctx.Err() != nil {
		res <- MetadataResult{song, ctx.Err()}
	}
	res <- MetadataResult{song, err}
}

func GetMetadata(query string) (*internal.Song, error) {
	firstArg := query
	if !strings.HasPrefix(query, config.LINK_PREFIX) {
		firstArg = fmt.Sprintf("ytsearch1:%s", strings.ReplaceAll(query, "\"", ""))
	}
	args := []string{
		firstArg,
		"-f", "bestaudio",
		"--max-downloads", "1",
		"--no-playlist",
		"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
		"--skip-download",
		"--print-json",
	}
	//log.Printf("yt-dlp %s", strings.Join(args, " "))
	var commandPath = "yt-dlp"
	if config.Utils != "" {
		commandPath = config.Utils + "yt-dlp.exe"
	}
	cmd := exec.Command(commandPath, args...)
	if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
		return nil, errors.Errorf("failed to get metadata of song (query %s): %v", query, err)
	} else {
		videoMetadata := internal.VideoMetadata{}
		err = json.Unmarshal(data, &videoMetadata)
		if err != nil {
			return nil, errors.Errorf("failed to unmarshal video metadata (query %s): %v", query, err)
		}
		return &internal.Song{
			Title:        videoMetadata.Title,
			Author:       videoMetadata.Uploader,
			Duration:     videoMetadata.Duration,
			OriginalUrl:  videoMetadata.OriginalURL,
			ThumbnailUrl: videoMetadata.Thumbnail,
			FileUrl:      videoMetadata.URL,
			Query:        query,
		}, nil
	}
}

func Search(query string, cnt int) ([]string, []string, error) {
	args := []string{
		fmt.Sprintf("ytsearch%d:%s", cnt, strings.ReplaceAll(query, "\"", "")),
		//"-f", "bestaudio",
		"--max-downloads", strconv.Itoa(cnt),
		"--no-playlist",
		"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
		"--skip-download",
		"--print-json",
	}
	var commandPath = "yt-dlp"
	if config.Utils != "" {
		commandPath = config.Utils + "yt-dlp.exe"
	}
	cmd := exec.Command(commandPath, args...)
	if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
		return nil, nil, errors.Errorf("failed to search song(%s): %v", query, err)
	} else {
		dec := json.NewDecoder(bytes.NewReader(data))
		videosMetadata := make([]internal.VideoMetadata, 0, cnt)
		for i := 0; i < cnt; i++ {
			metadata := internal.VideoMetadata{}
			if err = dec.Decode(&metadata); err == io.EOF {
				break
			} else if err != nil {
				continue
			}
			videosMetadata = append(videosMetadata, metadata)
		}
		titles := make([]string, len(videosMetadata))
		urls := make([]string, len(videosMetadata))
		for i, metadata := range videosMetadata {
			titles[i] = metadata.Title
			urls[i] = metadata.OriginalURL
		}
		return titles, urls, nil
	}
}
