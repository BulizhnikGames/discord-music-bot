package bot

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/go-faster/errors"
	"os"
)

func (voiceChat *VoiceEntity) InsertQueue(query string) {
	voiceChat.Queue.Write(internal.Song{Query: query})
}

func (bot *DiscordBot) ShuffleQueue(guildID string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.Queue.Shuffle()
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) SetLoop(guildID string, loop int) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if loop < 0 || loop > 2 {
			loop = 0
		}
		voiceChat.loop = loop
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) ClearQueue(guildID string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if voiceChat.nowPlaying != nil {
			save := voiceChat.loop
			voiceChat.loop = 0
			voiceChat.nowPlaying.Skip()
			voiceChat.loop = save
		}
		voiceChat.Queue.Clear()
		voiceChat.cache.Mutex.Lock()
		defer voiceChat.cache.Mutex.Unlock()
		os.Remove(config.Storage + guildID)
		clear(voiceChat.cache.Data)
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) GetQueue(guildID string) ([]string, error) {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if voiceChat.loop == 2 {
			if voiceChat.nowPlaying == nil {
				return []string{}, nil
			}
			return []string{fmt.Sprintf(
				"%s | %d:%02d by %s",
				voiceChat.nowPlaying.Title,
				voiceChat.nowPlaying.Duration/60,
				voiceChat.nowPlaying.Duration%60,
				voiceChat.nowPlaying.Author,
			)}, nil
		}
		res := make([]string, 0, voiceChat.Queue.Len+1)
		for song := range voiceChat.Queue.All() {
			var add string
			if song.Title != "" {
				add = fmt.Sprintf(
					"%s | %d:%02d by %s",
					song.Title,
					song.Duration/60,
					song.Duration%60,
					song.Author,
				)
			} else {
				add = song.Query
			}
			res = append(res, add)
		}
		if voiceChat.loop == 1 && voiceChat.nowPlaying != nil {
			res = append(
				res,
				fmt.Sprintf(
					"%s | %d:%02d by %s",
					voiceChat.nowPlaying.Title,
					voiceChat.nowPlaying.Duration/60,
					voiceChat.nowPlaying.Duration%60,
					voiceChat.nowPlaying.Author,
				),
			)
		}
		return res, nil
	} else {
		return nil, errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) SkipSong(guildID string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.Skip != nil {
			voiceChat.nowPlaying.Skip()
		}
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) Pause(guildID string, pause bool) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if voiceChat.nowPlaying != nil && &voiceChat.nowPlaying.Stream != nil {
			voiceChat.nowPlaying.Stream.SetPaused(pause)
		}
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) NowPlaying(guildID string) (*internal.Song, error) {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		return voiceChat.nowPlaying.Song, nil
	} else {
		return nil, errors.New("bot isn't in the voice chat")
	}
}
