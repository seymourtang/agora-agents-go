package core

import Agora "github.com/AgoraIO/agora-agents-go/v2"

type Profile string

const (
	ProfileGlobal Profile = "global"
	ProfileCN     Profile = "cn"
)

// TurnDetectionConfig configures conversation turn detection.
//
// AgentKit owns this type rather than aliasing the generated SDK struct so the
// agentkit-specific `language` field (which drives ASR language; see
// resolveTurnDetectionConfig) survives SDK regeneration. The remaining fields
// mirror Agora.StartAgentsRequestPropertiesTurnDetection. AgentKit serializes
// this config to a JSON map (see StructToMap) when building the join payload, so
// only JSON tags are required.
type TurnDetectionConfig struct {
	// Language is the BCP-47 language tag identifying the primary language used
	// for agent interaction. AgentKit-only field; not present in the generated SDK.
	Language            *Agora.AsrLanguage                                            `json:"language,omitempty"`
	Mode                *string                                                       `json:"mode,omitempty"`
	Config              *Agora.StartAgentsRequestPropertiesTurnDetectionConfig        `json:"config,omitempty"`
	Type                *Agora.StartAgentsRequestPropertiesTurnDetectionType          `json:"type,omitempty"`
	InterruptMode       *Agora.StartAgentsRequestPropertiesTurnDetectionInterruptMode `json:"interrupt_mode,omitempty"`
	InterruptDurationMs *float64                                                      `json:"interrupt_duration_ms,omitempty"`
	InterruptKeywords   []string                                                      `json:"interrupt_keywords,omitempty"`
	PrefixPaddingMs     *int                                                          `json:"prefix_padding_ms,omitempty"`
	SilenceDurationMs   *int                                                          `json:"silence_duration_ms,omitempty"`
	Threshold           *float64                                                      `json:"threshold,omitempty"`
	CreateResponse      *bool                                                         `json:"create_response,omitempty"`
	InterruptResponse   *bool                                                         `json:"interrupt_response,omitempty"`
	Eagerness           *Agora.StartAgentsRequestPropertiesTurnDetectionEagerness     `json:"eagerness,omitempty"`
}
type SalConfig = Agora.StartAgentsRequestPropertiesSal
type SalMode = Agora.StartAgentsRequestPropertiesSalSalMode
type AdvancedFeatures = Agora.StartAgentsRequestPropertiesAdvancedFeatures
type SessionParams = Agora.StartAgentsRequestPropertiesParameters
type SessionParamsInput = SessionParams
type GeofenceConfig = Agora.StartAgentsRequestPropertiesGeofence
type RtcConfig = Agora.StartAgentsRequestPropertiesRtc
type FillerWordsConfig = Agora.StartAgentsRequestPropertiesFillerWords

type TurnDetectionNestedConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfig
type StartOfSpeechConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeech
type StartOfSpeechMode = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechMode
type StartOfSpeechVadConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechVadConfig
type StartOfSpeechKeywordsConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechKeywordsConfig
type StartOfSpeechDisabledConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechDisabledConfig
type StartOfSpeechDisabledConfigStrategy = Agora.StartAgentsRequestPropertiesTurnDetectionConfigStartOfSpeechDisabledConfigStrategy
type EndOfSpeechConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeech
type EndOfSpeechMode = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechMode
type EndOfSpeechVadConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechVadConfig
type EndOfSpeechSemanticConfig = Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechSemanticConfig
type InterruptionConfig = Agora.StartAgentsRequestPropertiesInterruption
type InterruptionMode = Agora.StartAgentsRequestPropertiesInterruptionMode

type TurnDetectionType = Agora.StartAgentsRequestPropertiesTurnDetectionType
type InterruptMode = Agora.StartAgentsRequestPropertiesTurnDetectionInterruptMode
type Eagerness = Agora.StartAgentsRequestPropertiesTurnDetectionEagerness

type LlmGreetingConfigs = map[string]interface{}
type LlmGreetingConfigsMode = string
type McpServersItem = map[string]interface{}

type SilenceConfig = Agora.StartAgentsRequestPropertiesParametersSilenceConfig
type SilenceAction = Agora.StartAgentsRequestPropertiesParametersSilenceConfigAction
type FarewellConfig = Agora.StartAgentsRequestPropertiesParametersFarewellConfig
type ParametersDataChannel = Agora.StartAgentsRequestPropertiesParametersDataChannel

type ParametersAudioScenario string

const (
	ParametersAudioScenarioDefault  ParametersAudioScenario = "default"
	ParametersAudioScenarioChorus   ParametersAudioScenario = "chorus"
	ParametersAudioScenarioAIServer ParametersAudioScenario = "aiserver"
)

type GeofenceArea = Agora.StartAgentsRequestPropertiesGeofenceArea
type GeofenceExcludeArea = Agora.StartAgentsRequestPropertiesGeofenceExcludeArea

type LlmConfig = Agora.Llm
type MllmConfig = Agora.Mllm
type MllmTurnDetectionConfig = Agora.MllmTurnDetection
type MllmTurnDetectionMode = Agora.MllmTurnDetectionMode
type AsrConfig = Agora.Asr
type SttConfig = AsrConfig
type LlmStyle = Agora.LlmStyle
type SttVendor = string
type MllmVendor = Agora.MllmVendor
type AvatarVendor = Agora.StartAgentsRequestPropertiesAvatarVendor
type AgentConfig = Agora.StartAgentsRequestProperties
type AgentConfigUpdate = Agora.UpdateAgentsRequestProperties
type SessionInfo = Agora.GetAgentsResponse
type SessionListResponse = Agora.ListAgentsResponse
type SessionSummary = Agora.ListAgentsResponseDataListItem
type SessionStatus = Agora.ListAgentsResponseDataListItemStatus
type ConversationHistory = Agora.GetHistoryAgentsResponse
type ConversationTurn = Agora.GetHistoryAgentsResponseContentsItem
type ConversationRole = Agora.GetHistoryAgentsResponseContentsItemRole
type ConversationTurns = Agora.GetTurnsAgentsResponse
type ConversationSessionTurn = Agora.GetTurnsAgentsResponseTurnsItem
type ThinkResponse = Agora.AgentThinkAgentManagementResponse
type ThinkOnListeningAction = Agora.AgentThinkAgentManagementRequestOnListeningAction
type ThinkOnThinkingAction = Agora.AgentThinkAgentManagementRequestOnThinkingAction
type ThinkOnSpeakingAction = Agora.AgentThinkAgentManagementRequestOnSpeakingAction
type SpeakPriority = Agora.SpeakAgentsRequestPriority
type Labels = map[string]string
type TtsConfig = Agora.Tts
type AvatarConfig = Agora.StartAgentsRequestPropertiesAvatar

type FillerWordsTrigger = Agora.StartAgentsRequestPropertiesFillerWordsTrigger
type FillerWordsTriggerFixedTimeConfig = Agora.StartAgentsRequestPropertiesFillerWordsTriggerFixedTimeConfig
type FillerWordsContent = Agora.StartAgentsRequestPropertiesFillerWordsContent
type FillerWordsContentStaticConfig = Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfig
type FillerWordsContentSelectionRule = Agora.StartAgentsRequestPropertiesFillerWordsContentStaticConfigSelectionRule
