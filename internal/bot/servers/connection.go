package servers

import (
	"context"
	"github.com/BulizhnikGames/discord-music-bot/internal"
	"github.com/BulizhnikGames/discord-music-bot/internal/bot/servers/voice"
	"github.com/BulizhnikGames/discord-music-bot/internal/errors"
	"sync"
	"time"
)

const VOICE_TIMEOUT = 30 * time.Minute

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
		Logger:          server.Logger,
		Cache: internal.AsyncMap[string, *internal.SongCache]{
			Data:  make(map[string]*internal.SongCache),
			Mutex: &sync.RWMutex{},
		},
	}
	queue.SetHandler(server.VoiceChat.DownloadSong)
	server.VoiceChat.Leave = stop

	updateTimeoutTimer := make(chan struct{}, 10)
	go server.LeaveAfterTimeout(updateTimeoutTimer)
	go server.VoiceChat.PlaySongs(ctx, server.Session, updateTimeoutTimer)

	return server.VoiceChat, nil
}

func (server *Server) LeaveAfterTimeout(updateTimer <-chan struct{}) {
	timer := time.NewTimer(VOICE_TIMEOUT)
	for {
		select {
		case <-timer.C:
			err := server.TryLeaveVoiceChat()
			if err != nil {
				server.Logger.Printf("Couldn't leave voice chat after reaching timeout: %v", err)
			}
			return
		case <-updateTimer:
			timer = time.NewTimer(VOICE_TIMEOUT)
		}
	}
}

// TryLeaveVoiceChat leaves voice chat if connected
func (server *Server) TryLeaveVoiceChat() error {
	if server.VoiceChat != nil {
		return server.LeaveVoiceChat()
	}
	return nil
}

// LeaveVoiceChat leaves voice chat if connected, otherwise throws error
func (server *Server) LeaveVoiceChat() error {
	if server.VoiceChat != nil {
		err := server.ClearQueue("⛔  left voice channel  ⛔")
		if err != nil {
			return err
		}

		server.VoiceChat.Mutex.Lock()
		if server.VoiceChat.NowPlaying != nil {
			server.VoiceChat.NowPlaying.EncodeSession.Cleanup()
		}
		server.VoiceChat.Mutex.Unlock()

		server.VoiceChat.Leave()

		_ = server.VoiceChat.VoiceConnection.Disconnect()

		server.VoiceChat = nil

		server.Logger.Printf("Left voice chat (guild ID: %s)", server.GuildID)
		return nil
	} else {
		return errors.New("bot isn't in the voice chat")
	}
}

func (server *Server) HandleLeave() {
	if server.VoiceChat != nil {
		_ = server.ClearQueue("⛔  left voice channel  ⛔")

		server.VoiceChat.Mutex.Lock()
		if server.VoiceChat.NowPlaying != nil {
			server.VoiceChat.NowPlaying.EncodeSession.Cleanup()
		}
		server.VoiceChat.Mutex.Unlock()

		server.VoiceChat.Leave()

		server.VoiceChat = nil

		server.Logger.Printf("Handled leave voice chat (guild ID: %s)", server.GuildID)
	}
}
