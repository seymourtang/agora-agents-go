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

// CredentialMode is the shared credential mode accepted by provider options.
type CredentialMode = core.CredentialMode

type (
	LLM    = core.LLMVendor
	TTS    = core.TTSVendor
	STT    = core.STTVendor
	MLLM   = core.MLLMVendor
	Avatar = core.AvatarVendorConfig
)

const (
	credentialModeManaged = core.CredentialModeManaged
	credentialModeBYOK    = core.CredentialModeBYOK
)
