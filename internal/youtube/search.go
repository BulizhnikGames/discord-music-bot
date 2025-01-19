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
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type MetadataResult struct {
	Data *internal.Song
	Err  error
}

func GetMetadataWithContext(ctx context.Context, query, guildID string, res chan<- MetadataResult) {
	song, err := GetMetadata(query, guildID, false)
	if ctx.Err() != nil {
		res <- MetadataResult{song, ctx.Err()}
	}
	res <- MetadataResult{song, err}
}

func GetMetadata(query, guildID string, tryCookies bool) (*internal.Song, error) {
	if !strings.HasPrefix(query, config.LINK_PREFIX) {
		query = fmt.Sprintf("ytsearch1:%s", strings.ReplaceAll(query, "\"", ""))
	}
	args := []string{
		"-f", "bestaudio",
		"--max-downloads", "1",
		"--no-playlist",
		"--quiet",
		//"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
		"--skip-download",
		"--print-json",
		query,
	}
	//log.Printf("yt-dlp %s", strings.Join(args, " "))
	var commandPath = "yt-dlp"
	if config.Utils != "" {
		commandPath = config.Utils + "yt-dlp.exe"
	}
	var cmd *exec.Cmd
	if tryCookies {
		cookiesArgs := []string{
			"--cookies", config.Cookies,
		}
		cookiesArgs = append(cookiesArgs, args...)
		cmd = exec.Command(commandPath, cookiesArgs...)
	} else {
		cmd = exec.Command(commandPath, args...)
	}
	log.Println(strings.Join(cmd.Args, " "))
	if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
		// Try cookies only if this try was without them
		if err.Error() == "exit status 1" && !tryCookies && config.Cookies != "" && config.CookiesGuildID == guildID {
			return GetMetadata(query, guildID, true)
		}
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
		"--max-downloads", strconv.Itoa(cnt),
		"--no-playlist",
		"--clean-info-json",
		"--quiet",
		"--match-filter", fmt.Sprintf("duration < %d & !is_live", 20*60),
		"--skip-download",
		"--print-json",
		fmt.Sprintf("ytsearch%d:%s", cnt, strings.ReplaceAll(query, "\"", "")),
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
