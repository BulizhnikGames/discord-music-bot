package voice

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"github.com/BulizhnikGames/discord-music-bot/internal/youtube"
	"log"
)

func (voiceChat *Connection) InsertQueue(song internal.Song) {
	voiceChat.Queue.Write(song)
}

func (voiceChat *Connection) DownloadSong(ctx context.Context, song *internal.Song) (*internal.Song, error) {
	voiceChat.Cache.Mutex.Lock()
	if cache, ok := voiceChat.Cache.Data[song.Query]; ok {
		log.Printf("Song is already in Cache: %s", song.Query)
		cache.Cnt++
		voiceChat.Cache.Mutex.Unlock()
		return cache.Song, nil
	}
	voiceChat.Cache.Mutex.Unlock()
	if song.HasAllData() {
		log.Printf("Song (%s) already has all metadata on it", song.Title)
		voiceChat.Cache.Mutex.Lock()
		defer voiceChat.Cache.Mutex.Unlock()
		voiceChat.Cache.Data[song.Query] = &internal.SongCache{
			Cnt:  1,
			Song: song,
		}
		return song, nil
	}
	log.Printf("Getting metadata of song: %s (%s)", song.Query, song.Title)
	res := make(chan youtube.MetadataResult)
	go youtube.GetMetadataWithContext(ctx, song.Query, voiceChat.VoiceConnection.GuildID, res)
	select {
	case data := <-res:
		if data.Err != nil {
			return nil, errors.Newf("couldn't get metadata of song: %v", data.Err)
		}
		log.Printf("Got metadata of song: %v", data.Data.Title)
		voiceChat.Cache.Mutex.Lock()
		defer voiceChat.Cache.Mutex.Unlock()
		voiceChat.Cache.Data[song.Query] = &internal.SongCache{
			Cnt:  1,
			Song: data.Data,
		}
		return data.Data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (voiceChat *Connection) GetVoiceChatID() string {
	return voiceChat.VoiceConnection.ChannelID
}
