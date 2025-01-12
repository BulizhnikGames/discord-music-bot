package internal

import (
	"github.com/jogramming/dca"
	"log"
	"os"
	"sync"
)

type Song struct {
	Title    string
	Author   string
	URL      string
	Query    string
	FilePath string
	Duration int
}

func (song *Song) Delete() {
	log.Printf("Delete song %s", song.Title)
	err := os.Remove(song.FilePath)
	if err != nil {
		log.Printf("Delete song file error: %s\n", err)
	}
}

type SongCache struct {
	Cnt int
	*Song
}

type PlayingSong struct {
	Skip   func(bool)
	Stream *dca.StreamingSession
	*Song
}

type AsyncMap[K comparable, V any] struct {
	Data  map[K]V
	Mutex *sync.RWMutex
}

func (m *AsyncMap[K, V]) Put(k K, v V) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Data[k] = v
}

func (m *AsyncMap[K, V]) Remove(k K) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	delete(m.Data, k)
}

type VideoMetadata struct {
	ID                   string      `json:"id"`
	Title                string      `json:"title"`
	Thumbnail            string      `json:"thumbnail"`
	Description          string      `json:"description"`
	Uploader             string      `json:"uploader"`
	UploaderID           string      `json:"uploader_id"`
	UploaderURL          string      `json:"uploader_url"`
	ChannelID            string      `json:"channel_id"`
	ChannelURL           string      `json:"channel_url"`
	Duration             int         `json:"duration"`
	ViewCount            int         `json:"view_count"`
	AverageRating        interface{} `json:"average_rating"`
	AgeLimit             int         `json:"age_limit"`
	WebpageURL           string      `json:"webpage_url"`
	Categories           []string    `json:"categories"`
	Tags                 []string    `json:"tags"`
	PlayableInEmbed      bool        `json:"playable_in_embed"`
	LiveStatus           interface{} `json:"live_status"`
	ReleaseTimestamp     interface{} `json:"release_timestamp"`
	CommentCount         interface{} `json:"comment_count"`
	LikeCount            int         `json:"like_count"`
	Channel              string      `json:"channel"`
	ChannelFollowerCount int         `json:"channel_follower_count"`
	UploadDate           string      `json:"upload_date"`
	Availability         string      `json:"availability"`
	OriginalURL          string      `json:"original_url"`
	WebpageURLBasename   string      `json:"webpage_url_basename"`
	WebpageURLDomain     string      `json:"webpage_url_domain"`
	Extractor            string      `json:"extractor"`
	ExtractorKey         string      `json:"extractor_key"`
	PlaylistCount        int         `json:"playlist_count"`
	Playlist             string      `json:"playlist"`
	PlaylistID           string      `json:"playlist_id"`
	PlaylistTitle        string      `json:"playlist_title"`
	PlaylistUploader     interface{} `json:"playlist_uploader"`
	PlaylistUploaderID   interface{} `json:"playlist_uploader_id"`
	NEntries             int         `json:"n_entries"`
	PlaylistIndex        int         `json:"playlist_index"`
	LastPlaylistIndex    int         `json:"__last_playlist_index"`
	PlaylistAutonumber   int         `json:"playlist_autonumber"`
	DisplayID            string      `json:"display_id"`
	Fulltitle            string      `json:"fulltitle"`
	DurationString       string      `json:"duration_string"`
	RequestedSubtitles   interface{} `json:"requested_subtitles"`
	Asr                  int         `json:"asr"`
	Filesize             int         `json:"filesize"`
	FormatID             string      `json:"format_id"`
	FormatNote           string      `json:"format_note"`
	SourcePreference     int         `json:"source_preference"`
	Fps                  interface{} `json:"fps"`
	AudioChannels        int         `json:"audio_channels"`
	Height               interface{} `json:"height"`
	HasDrm               bool        `json:"has_drm"`
	Tbr                  float64     `json:"tbr"`
	URL                  string      `json:"url"`
	Width                interface{} `json:"width"`
	Language             string      `json:"language"`
	LanguagePreference   int         `json:"language_preference"`
	Preference           interface{} `json:"preference"`
	Ext                  string      `json:"ext"`
	Vcodec               string      `json:"vcodec"`
	Acodec               string      `json:"acodec"`
	DynamicRange         interface{} `json:"dynamic_range"`
	Abr                  float64     `json:"abr"`
	Filename             string      `json:"filename"`
}
