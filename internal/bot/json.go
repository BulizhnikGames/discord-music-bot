package bot

import "strings"

func (voiceChat *VoiceEntity) constructJsonLine(comps ...string) []byte {
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

func (voiceChat *VoiceEntity) pauseButtonJson() string {
	pause := false
	if voiceChat.nowPlaying != nil && voiceChat.nowPlaying.Stream != nil {
		pause = voiceChat.nowPlaying.Stream.Paused()
	}
	if pause {
		return `{
          "custom_id": "resume",
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
          "custom_id": "pause",
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

func (voiceChat *VoiceEntity) skipButtonJson() string {
	return `{
          "custom_id": "skip",
          "type": 2,
          "style": 1,
          "label": "Skip",
          "emoji": {
            "name": "⏩",
            "animated": false
          }
        }`
}

func (voiceChat *VoiceEntity) stopButtonJson() string {
	return `{
          "custom_id": "clear",
          "type": 2,
          "style": 4,
          "label": "Stop",
          "emoji": {
            "name": "⏹️",
            "animated": false
          }
        }`
}

func (voiceChat *VoiceEntity) shuffleQueueJson() string {
	return `{
          "custom_id": "shuffle",
          "type": 2,
          "style": 2,
          "label": "Shuffle queue",
          "emoji": {
            "name": "🔀",
            "animated": false
          }
        }`
}

func (voiceChat *VoiceEntity) loopOptsJson() string {
	switch voiceChat.loop {
	case 0:
		return `{
          "custom_id": "loop1",
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
          "custom_id": "loop2",
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
          "custom_id": "loop0",
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
          "custom_id": "loop1",
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
