package internal

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jogramming/dca"
	"sync"
)

type Song struct {
	Title        string
	Author       string
	Duration     int
	OriginalUrl  string
	FileUrl      string
	ThumbnailUrl string
	Query        string
}

// HasAllData is true only if all Song's fields have some value
func (s Song) HasAllData() bool {
	return len(s.Title)*len(s.Author)*s.Duration*len(s.OriginalUrl)*len(s.FileUrl)*len(s.ThumbnailUrl) > 0
}

type SongCache struct {
	Cnt int
	*Song
}

type PlayingSong struct {
	*Song
	// Skip playing song, won't stop playback when playback is looped over song
	Skip func(playbackText string)
	// Playback message
	Message       *discordgo.Message
	Stream        *dca.StreamingSession
	EncodeSession *dca.EncodeSession
}

type AsyncMap[K comparable, V any] struct {
	Data  map[K]V
	Mutex *sync.RWMutex
}

type VideoMetadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Thumbnail   string `json:"thumbnail"`
	Uploader    string `json:"uploader"`
	Duration    int    `json:"duration"`
	OriginalURL string `json:"original_url"`
	URL         string `json:"url"`
}
