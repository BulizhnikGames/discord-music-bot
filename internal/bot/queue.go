package bot

import (
	"fmt"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/go-faster/errors"
)

func (voiceChat *VoiceEntity) InsertQueue(query string) {
	voiceChat.Queue.Write(&internal.Song{Query: query})
}

/*func (voiceChat *VoiceEntity) insertToSongQueue(song internal.SongCache) {
	voiceChat.queueMutex.Lock()
	defer voiceChat.queueMutex.Unlock()
	voiceChat.songQueue <- song
}*/

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
		voiceChat.Queue.Clear()
		voiceChat.cache.Mutex.Lock()
		defer voiceChat.cache.Mutex.Unlock()
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
		songs := voiceChat.Queue.Get()
		res := make([]string, 0, len(songs))
		for _, song := range songs {
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
		return res, nil
	} else {
		return nil, errors.New("bot isn't in the voice chat")
	}
}

func (bot *DiscordBot) SkipSong(guildID string) error {
	bot.VoiceEntities.Mutex.RLock()
	defer bot.VoiceEntities.Mutex.RUnlock()
	if voiceChat, ok := bot.VoiceEntities.Data[guildID]; ok {
		if voiceChat.skip != nil {
			voiceChat.skip()
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
		return voiceChat.nowPlaying, nil
	} else {
		return nil, errors.New("bot isn't in the voice chat")
	}
}
