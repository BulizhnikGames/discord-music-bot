package servers

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"log"
	"os"
	"time"
)

const SEC = time.Second

func (server *Server) ShuffleQueue() error {
	if server.VoiceChat != nil {
		server.VoiceChat.Queue.Shuffle()
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (server *Server) SetLoop(loop int) error {
	if server.VoiceChat != nil {
		if loop < 0 || loop > 2 {
			loop = 0
		}
		server.VoiceChat.Mutex.Lock()
		server.VoiceChat.Loop = loop
		server.VoiceChat.Mutex.Unlock()
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (server *Server) ClearQueue(playbackText string) error {
	if server.VoiceChat != nil {
		server.VoiceChat.Queue.Clear()

		server.VoiceChat.ForceStop(playbackText)

		server.VoiceChat.Cache.Mutex.Lock()
		clear(server.VoiceChat.Cache.Data)
		server.VoiceChat.Cache.Mutex.Unlock()

		err := os.RemoveAll(config.Storage + server.GuildID + "/")
		if err != nil {
			log.Printf("couldn't delete guilds Cache: %v", err)
		}

		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (server *Server) GetQueue() ([]string, error) {
	if server.VoiceChat != nil {
		server.VoiceChat.Mutex.RLock()
		loop, nowPlaying := server.VoiceChat.Loop, server.VoiceChat.NowPlaying
		server.VoiceChat.Mutex.RUnlock()
		if loop == 2 {
			if nowPlaying == nil {
				return []string{}, nil
			}
			return []string{nowPlaying.Title}, nil
		}
		res := make([]string, 0, server.VoiceChat.Queue.Len+1)
		for song := range server.VoiceChat.Queue.All() {
			if song.Title != "" && song.Duration != 0 && song.Author != "" {
				res = append(res, fmt.Sprintf(
					"%s - [%d:%02d] by %s",
					song.Title,
					song.Duration/60,
					song.Duration%60,
					song.Author,
				))
			} else {
				res = append(res, song.Query)
			}
		}
		if loop == 1 && nowPlaying != nil {
			res = append(res, nowPlaying.Title)
		}
		return res, nil
	} else {
		return nil, errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (server *Server) SkipSong(playbackText string) error {
	if server.VoiceChat != nil {
		server.VoiceChat.Mutex.RLock()
		if server.VoiceChat.NowPlaying != nil && server.VoiceChat.NowPlaying.Skip != nil {
			skip := server.VoiceChat.NowPlaying.Skip
			server.VoiceChat.Mutex.RUnlock()
			skip(playbackText)
			return nil
		}
		server.VoiceChat.Mutex.RUnlock()
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (server *Server) Pause(pause bool) error {
	if server.VoiceChat != nil {
		server.VoiceChat.Mutex.RLock()
		defer server.VoiceChat.Mutex.RUnlock()
		if server.VoiceChat.NowPlaying != nil && &server.VoiceChat.NowPlaying.Stream != nil {
			server.VoiceChat.NowPlaying.Stream.SetPaused(pause)
		}
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (server *Server) NowPlaying() (*internal.Song, int, error) {
	if server.VoiceChat != nil {
		server.VoiceChat.Mutex.RLock()
		defer server.VoiceChat.Mutex.RUnlock()
		if server.VoiceChat.NowPlaying == nil || server.VoiceChat.NowPlaying.Stream == nil {
			return nil, 0, nil
		}
		return server.VoiceChat.NowPlaying.Song, int(server.VoiceChat.NowPlaying.Stream.PlaybackPosition() / SEC), nil
	} else {
		return nil, 0, errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}
