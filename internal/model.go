package internal

import (
	"context"
	"github.com/jogramming/dca"
	"log"
	"math/rand/v2"
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
	Skip   func()
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

type CycleQueue[T any] struct {
	readIdx, writeIdx int
	buffer            []*T
	mutex             *sync.RWMutex
	handled           AsyncMap[*T, bool]
	pos               AsyncMap[*T, int]
	ctx               context.Context
	WriteHandler      func(ctx context.Context, val *T) (*T, error)
	stopHandlers      context.CancelFunc
}

func CreateCycleQueue[T any](size int) *CycleQueue[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &CycleQueue[T]{
		buffer: make([]*T, size),
		mutex:  &sync.RWMutex{},
		handled: AsyncMap[*T, bool]{
			Data:  make(map[*T]bool),
			Mutex: &sync.RWMutex{},
		},
		pos: AsyncMap[*T, int]{
			Data:  make(map[*T]int),
			Mutex: &sync.RWMutex{},
		},
		ctx:          ctx,
		stopHandlers: cancel,
	}
}

func (queue *CycleQueue[T]) SetHandler(handler func(ctx context.Context, val *T) (*T, error)) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.WriteHandler = handler
	//log.Printf("Setted handler: %v", handler)
}

func (queue *CycleQueue[T]) Len() int {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()
	return (queue.writeIdx - queue.readIdx + len(queue.buffer)) % len(queue.buffer)
}

func (queue *CycleQueue[T]) Cap() int {
	return len(queue.buffer)
}

func (queue *CycleQueue[T]) Write(v *T) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.buffer == nil || queue.buffer[queue.writeIdx] != nil {
		return
	}
	//log.Printf("write")
	//writtenAt := queue.writeIdx
	queue.pos.Mutex.Lock()
	defer queue.pos.Mutex.Unlock()
	queue.pos.Data[v] = queue.writeIdx
	queue.buffer[queue.writeIdx] = v
	queue.writeIdx = (queue.writeIdx + 1) % len(queue.buffer)
	//log.Printf("written")
	//log.Printf("handler: %v", queue.WriteHandler)
	if queue.WriteHandler != nil {
		//log.Printf("handle")
		go func() {
			processed, err := queue.WriteHandler(queue.ctx, v)
			if err != nil {
				log.Printf("couldn't handle new element: %v", err)
			}
			queue.mutex.Lock()
			queue.handled.Mutex.Lock()
			queue.pos.Mutex.Lock()
			defer queue.mutex.Unlock()
			defer queue.handled.Mutex.Unlock()
			defer queue.pos.Mutex.Unlock()
			queue.buffer[queue.pos.Data[v]] = processed
			queue.pos.Data[processed] = queue.pos.Data[v]
			delete(queue.pos.Data, v)
			queue.handled.Data[processed] = true
		}()
	}
}

func (queue *CycleQueue[T]) Read() *T {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.buffer == nil || queue.buffer[queue.readIdx] == nil {
		return nil
	}
	v := queue.buffer[queue.readIdx]
	queue.buffer[queue.readIdx] = nil
	queue.readIdx = (queue.readIdx + 1) % len(queue.buffer)
	return v
}

func (queue *CycleQueue[T]) ReadHandled() *T {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.buffer == nil || queue.buffer[queue.readIdx] == nil {
		return nil
	}
	v := queue.buffer[queue.readIdx]
	queue.handled.Mutex.RLock()
	defer queue.handled.Mutex.RUnlock()
	if handled, ok := queue.handled.Data[v]; !ok || !handled {
		return nil
	}
	queue.buffer[queue.readIdx] = nil
	queue.readIdx = (queue.readIdx + 1) % len(queue.buffer)
	return v
}

func (queue *CycleQueue[T]) Clear() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	clear(queue.buffer)
	queue.stopHandlers()
	queue.readIdx = 0
	queue.writeIdx = 0
	queue.ctx, queue.stopHandlers = context.WithCancel(context.Background())
}

func (queue *CycleQueue[T]) Get() []T {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	res := make([]T, 0, (queue.writeIdx-queue.readIdx+len(queue.buffer))%len(queue.buffer))
	save := queue.readIdx
	for queue.buffer[queue.readIdx] != nil {
		res = append(res, *queue.buffer[queue.readIdx])
		queue.readIdx = (queue.readIdx + 1) % len(queue.buffer)
	}
	queue.readIdx = save
	return res
}

func (queue *CycleQueue[T]) Shuffle() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	vals := make([]*T, 0, (queue.writeIdx-queue.readIdx+len(queue.buffer))%len(queue.buffer))
	save := queue.readIdx
	for queue.buffer[queue.readIdx] != nil {
		vals = append(vals, queue.buffer[queue.readIdx])
		queue.readIdx = (queue.readIdx + 1) % len(queue.buffer)
	}
	queue.readIdx = save
	rand.Shuffle(len(vals), func(i, j int) {
		vals[i], vals[j] = vals[j], vals[i]
	})
	queue.writeIdx = (queue.writeIdx - len(vals) + len(queue.buffer)) % len(queue.buffer)
	for _, val := range vals {
		queue.pos.Mutex.Lock()
		queue.pos.Data[val] = queue.writeIdx
		queue.buffer[queue.writeIdx] = val
		queue.writeIdx = (queue.writeIdx + 1) % len(queue.buffer)
		queue.pos.Mutex.Unlock()
	}
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
