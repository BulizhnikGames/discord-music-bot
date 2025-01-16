package servers

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers/voice"
	"github.com/BulizhnikGames/discord-music-bot/internal/config"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"log"
	"os"
	"sync"
)

// JoinVoiceChat bot joins voice chat if it isn't already in one.
// If bot is already in voice chat:
// 1. Returns error if he's in other voice chat
// 2. Returns queue if he's in requested voice chat
func (server *Server) JoinVoiceChat(guildID, voiceChannel, textChannel string) (*voice.Connection, error) {
	if server.VoiceChat != nil {
		if server.VoiceChat.VoiceConnection.ChannelID == voiceChannel {
			return server.VoiceChat, nil
		} else {
			err := errors.Newf(
				"could not join voiceConnection chat (id: %s, guild: %s): already in another channel",
				server.VoiceChat.VoiceConnection.ChannelID,
				guildID,
			)
			return nil, err.AddUser("Bot is already in other channel")
		}
	}

	voiceConn, err := server.Session.ChannelVoiceJoin(guildID, voiceChannel, false, true)
	if err != nil {
		err := errors.Newf("could not join voiceConnection chat (id: %s, guild: %s): %v", voiceChannel, guildID, err)
		return nil, err.AddUser("Couldn't join voice chat")
	}

	ctx, stop := context.WithCancel(context.Background())
	queue := internal.CreateCycleQueue(ctx, 150)
	go queue.Run()
	server.VoiceChat = &voice.Connection{
		VoiceConnection: voiceConn,
		Queue:           queue,
		TextChannel:     textChannel,
		Mutex:           &sync.RWMutex{},
		Cache: internal.AsyncMap[string, *internal.SongCache]{
			Data:  make(map[string]*internal.SongCache),
			Mutex: &sync.RWMutex{},
		},
	}
	queue.SetHandler(server.VoiceChat.DownloadSong)
	server.VoiceChat.Leave = stop

	go server.VoiceChat.PlaySongs(ctx, server.Session)

	return server.VoiceChat, nil
}

func (server *Server) LeaveVoiceChat() error {
	if server.VoiceChat != nil {
		err := server.VoiceChat.VoiceConnection.Disconnect()
		if err != nil {
			return err
		}
		server.VoiceChat = nil

		err = os.RemoveAll(config.Storage + server.GuildID + "/")
		if err != nil {
			log.Printf("couldn't delete guilds cache: %v", err)
		}
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}
