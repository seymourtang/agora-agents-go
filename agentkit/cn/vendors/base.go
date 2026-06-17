package vendors

import core "github.com/AgoraIO/agora-agents-go/v2/agentkit/core"

type SampleRate = core.SampleRate

const (
	SampleRate8kHz  = core.SampleRate8kHz
	SampleRate16kHz = core.SampleRate16kHz
	SampleRate22kHz = core.SampleRate22kHz
	SampleRate24kHz = core.SampleRate24kHz
	SampleRate44kHz = core.SampleRate44kHz
	SampleRate48kHz = core.SampleRate48kHz
)

type LLM interface {
	ToConfig() map[string]interface{}
}

type TTS interface {
	ToConfig() map[string]interface{}
	GetSampleRate() *SampleRate
}

type STT interface {
	ToConfig() map[string]interface{}
}

type Avatar interface {
	ToConfig() map[string]interface{}
	RequiredSampleRate() SampleRate
}
