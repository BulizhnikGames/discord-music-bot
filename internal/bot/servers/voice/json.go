package voice

import (
	"strings"
)

func ConstructJsonLine(comps ...string) []byte {
	if len(comps) == 0 {
		return []byte{}
	}
	res := strings.Builder{}
	res.WriteString(`{ "type": 1, "components": [ `)
	for i, comp := range comps {
		res.WriteString(comp)
		if i < len(comps)-1 {
			res.WriteString(", ")
		}
	}
	res.WriteString(` ] }`)
	return []byte(res.String())
}

func (voiceChat *Connection) pauseButtonJson(sessionID string) string {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()
	pause := false
	if voiceChat.NowPlaying != nil && voiceChat.NowPlaying.Stream != nil {
		pause = voiceChat.NowPlaying.Stream.Paused()
	}
	if pause {
		return `{
          "custom_id": "` + sessionID + `:resume",
          "type": 2,
          "style": 3,
          "label": "Resume",
          "emoji": {
            "name": "▶️",
            "animated": false
          }
        }`
	} else {
		return `{
          "custom_id": "` + sessionID + `:pause",
          "type": 2,
          "style": 2,
          "label": "Pause",
          "emoji": {
            "name": "⏸️",
            "animated": false
          }
        }`
	}
}

func (voiceChat *Connection) skipButtonJson(sessionID string) string {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()
	return `{
          "custom_id": "` + sessionID + `:skip",
          "type": 2,
          "style": 1,
          "label": "Skip",
          "emoji": {
            "name": "⏩",
            "animated": false
          }
        }`
}

func (voiceChat *Connection) stopButtonJson(sessionID string) string {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()
	return `{
          "custom_id": "` + sessionID + `:clear",
          "type": 2,
          "style": 4,
          "label": "Stop",
          "emoji": {
            "name": "⏹️",
            "animated": false
          }
        }`
}

func (voiceChat *Connection) shuffleQueueJson(sessionID string) string {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()
	return `{
          "custom_id": "` + sessionID + `:shuffle",
          "type": 2,
          "style": 2,
          "label": "Shuffle Queue",
          "emoji": {
            "name": "🔀",
            "animated": false
          }
        }`
}

func (voiceChat *Connection) loopOptsJson(sessionID string) string {
	voiceChat.Mutex.RLock()
	defer voiceChat.Mutex.RUnlock()
	switch voiceChat.Loop {
	case 0:
		return `{
          "custom_id": "` + sessionID + `:loop1",
          "type": 2,
          "style": 2,
          "label": "Loop Queue",
          "emoji": {
            "name": "🔁",
            "animated": false
          }
        }`
	case 1:
		return `{
          "custom_id": "` + sessionID + `:loop2",
          "type": 2,
          "style": 2,
          "label": "Loop Song",
          "emoji": {
            "name": "🔂",
            "animated": false
          }
        }`
	case 2:
		return `{
          "custom_id": "` + sessionID + `:loop0",
          "type": 2,
          "style": 2,
          "label": "No Loop",
          "emoji": {
            "name": "↪️",
            "animated": false
          }
        }`
	default:
		return `{
          "custom_id": "` + sessionID + `:loop1",
          "type": 2,
          "style": 2,
          "label": "Loop Queue",
          "emoji": {
            "name": "🔁",
            "animated": false
          }
        }`
	}
}
