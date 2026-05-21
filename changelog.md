# Changelog

## v1.5.0

AgentKit alignment for Conversational AI v2.7.

### Added

- Added `AgoraClient.Telephony` and `AgoraClient.PhoneNumbers` so AgentKit callers can reach the v2.7 telephony and phone-number REST endpoints without rebuilding the generated client.
- Added `vendors.NewXaiGrok` for xAI Grok MLLM sessions (`mllm.vendor`: `"xai"`), matching the TypeScript `XaiGrok` shape. `NewXAIGrok` remains as a deprecated alias.
- Added `vendors.NewGenericAvatar` and `IsGenericAvatar`.
- Added avatar parameter enrichment: Generic avatars get `agora_appid`, `agora_channel`, and `agora_token` from the session when omitted; LiveAvatar and HeyGen get `agora_token` auto-generated when omitted.
- Added `GenerateAvatarRtcToken` for advanced avatar token generation. Avatar tokens use the same ConvoAI token format as agent tokens and are scoped to the avatar UID.
- Added `WithGreetingConfigs` for `llm.greeting_configs`, including v2.7 `interruptable`.
- Added `GetTurnsOptions` and `GetAllTurns` for turn pagination. `GetAllTurns` returns the full response with aggregated `Turns`.
- Added `ThinkOnListeningAction*`, `ThinkOnThinkingAction*`, and `ThinkOnSpeakingAction*` constants for v2.7 Think actions.
- Added `InterruptionModeStartOfSpeech`, `InterruptionModeKeywords`, `InterruptionDisabledStrategyAppend`, and `InterruptionDisabledStrategyIgnore` convenience constants for the v2.7 `interruption` object.
- Added `SpeakPriorityInterrupt`, `SpeakPriorityAppend`, and `SpeakPriorityIgnore` convenience constants for `AgentSession.Say`.
- Added `MllmTurnDetectionModeAgoraVad`, `MllmTurnDetectionModeServerVad`, and `MllmTurnDetectionModeSemanticVad` convenience constants for the MLLM `turn_detection.mode` field.
- Added `AzureOpenAIOptions.Model` so the underlying deployment model is emitted as `params.model` for parity with the TypeScript SDK; Azure ignores the value for chat completions, but downstream tooling and logs surface it.

### Changed

- `AgentSession.Start` now sends a map-based join payload after preset resolution, preventing generated structs from reintroducing empty provider-owned fields such as `llm.url`, `llm.api_key`, or `tts.params.key`.
- `ToPropertiesMap` now builds vendor configs from maps directly for closer parity with Python and TypeScript AgentKit.
- `GetTurns` supports `page_index` and `page_size`; callers with more than one page should paginate or call `GetAllTurns`.
- `Agent.ToPropertiesMap` now rejects MLLM + enabled avatar combinations before generating tokens or building properties; avatars currently require the cascading ASR/LLM/TTS pipeline.
- Avatar vendor `ToConfig()` (HeyGen, LiveAvatar, Akool, Anam, Generic) now spreads `AdditionalParams` first so required fields like `api_key`, `quality`, and `agora_uid` always take precedence over caller overrides.
- `OpenAIRealtime.ToConfig` now lets explicit `Params["model"]` override the named `Model`, matching the TypeScript SDK and the existing Gemini/Vertex AI/xAI Grok behavior.

### Documentation

- Documented v2.7 error `reason` codes, Think default behavior, event `112`, avatar token handling, Generic avatars, xAI Grok, Deepgram TTS, Murf, Anam, LiveAvatar, `pause_state_enabled`, `audio_scenario`, and `pipeline_id`.

### Migration Notes

- In v2.7, omitting `ThinkOptions.OnListeningAction` uses the server default `interrupt`. Pass `agentkit.ThinkOnListeningActionInject.Ptr()` to preserve inject-style behavior.
- Avatar `AgoraUID` should be distinct from the session `AgentUID`. The SDK warns on collisions and preserves explicitly provided avatar tokens.
