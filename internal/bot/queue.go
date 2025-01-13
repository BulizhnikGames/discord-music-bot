package bot

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"log"
	"os"
	"time"
)

func (voiceChat *VoiceEntity) InsertQueue(song internal.Song) {
	voiceChat.queue.Write(song)
}

func (bot *DiscordBot) ShuffleQueue(guildID string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.queue.Shuffle()
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) SetLoop(guildID string, loop int) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if loop < 0 || loop > 2 {
			loop = 0
		}
		voiceChat.mutex.Lock()
		voiceChat.loop = loop
		voiceChat.mutex.Unlock()
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) ClearQueue(guildID string, playbackText string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.queue.Clear()

		voiceChat.forceStop(playbackText)

		voiceChat.cache.Mutex.Lock()
		clear(voiceChat.cache.Data)
		voiceChat.cache.Mutex.Unlock()

		err := os.RemoveAll(config.Storage + guildID + "/")
		if err != nil {
			log.Printf("couldn't delete guilds cache: %v", err)
		}

		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) GetQueue(guildID string) ([]string, error) {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.mutex.RLock()
		loop, nowPlaying := voiceChat.loop, voiceChat.nowPlaying
		voiceChat.mutex.RUnlock()
		if loop == 2 {
			if nowPlaying == nil {
				return []string{}, nil
			}
			return []string{fmt.Sprintf(
				"%s | %d:%02d by %s",
				nowPlaying.Title,
				nowPlaying.Duration/60,
				nowPlaying.Duration%60,
				nowPlaying.Author,
			)}, nil
		}
		res := make([]string, 0, voiceChat.queue.Len+1)
		for song := range voiceChat.queue.Part(10) {
			var add string
			if song.Title != "" && song.Duration != 0 && song.Author != "" {
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
		if loop == 1 && nowPlaying != nil {
			res = append(
				res,
				fmt.Sprintf(
					"%s | %d:%02d by %s",
					nowPlaying.Title,
					nowPlaying.Duration/60,
					nowPlaying.Duration%60,
					nowPlaying.Author,
				),
			)
		}
		return res, nil
	} else {
		return nil, errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) SkipSong(guildID string, playbackText string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.mutex.RLock()
		if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.Skip != nil {
			skip := voiceChat.nowPlaying.Skip
			voiceChat.mutex.RUnlock()
			skip(playbackText)
			return nil
		}
		voiceChat.mutex.RUnlock()
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) Pause(guildID string, pause bool) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.mutex.RLock()
		defer voiceChat.mutex.RUnlock()
		if voiceChat.nowPlaying != nil && &voiceChat.nowPlaying.Stream != nil {
			voiceChat.nowPlaying.Stream.SetPaused(pause)
		}
		return nil
	} else {
		return errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) NowPlaying(guildID string) (*internal.Song, int, error) {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		voiceChat.mutex.RLock()
		defer voiceChat.mutex.RUnlock()
		if voiceChat.nowPlaying == nil || voiceChat.nowPlaying.Stream == nil {
			return nil, 0, nil
		}
		return voiceChat.nowPlaying.Song, int(voiceChat.nowPlaying.Stream.PlaybackPosition() / time.Second), nil
	} else {
		return nil, 0, errors.New("Bot isn't in the voice chat").AddUser("Bot isn't in the voice chat")
	}
}
