package agentkit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	Agora "github.com/AgoraIO-Conversational-AI/agent-server-sdk-go"
	"github.com/AgoraIO-Conversational-AI/agent-server-sdk-go/agentkit/vendors"
	"github.com/AgoraIO-Conversational-AI/agent-server-sdk-go/client"
	"github.com/AgoraIO-Conversational-AI/agent-server-sdk-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type disabledMllmVendor struct{}

func (disabledMllmVendor) ToConfig() map[string]interface{} {
	return map[string]interface{}{
		"vendor": "openai",
		"enable": false,
	}
}

func TestToPropertiesSupportsPresetFlowAndRTMDefault(t *testing.T) {
	enableRTM := true
	agent := NewAgent(
		WithInstructions("Preset flow"),
		WithAdvancedFeatures(&AdvancedFeatures{EnableRtm: &enableRTM}),
	)

	props, err := agent.ToProperties(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)
	require.NotNil(t, props)
	assert.Equal(t, "room-1", props.Channel)
	assert.NotNil(t, props.Parameters)
	require.NotNil(t, props.Parameters.DataChannel)
	assert.Equal(t, "rtm", string(*props.Parameters.DataChannel))
	assert.Nil(t, props.Llm)
	assert.Nil(t, props.Tts)
}

func TestToPropertiesMapIncludesAudioScenario(t *testing.T) {
	agent := NewAgent(WithAudioScenario(ParametersAudioScenarioAIServer))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	parameters, ok := props["parameters"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "aiserver", parameters["audio_scenario"])
}

func TestWithToolsSetsEnableTools(t *testing.T) {
	props, err := NewAgent().WithTools(true).ToProperties(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)
	require.NotNil(t, props.AdvancedFeatures)
	require.NotNil(t, props.AdvancedFeatures.EnableTools)
	assert.True(t, *props.AdvancedFeatures.EnableTools)
}

func TestCreateSessionStartIncludesPresetPipelineAndGetTurns(t *testing.T) {
	var started int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/projects/appid/join":
			var req map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "support-agent", req["name"])
			assert.Equal(t, "deepgram_nova_3,openai_gpt_4o_mini,openai_tts_1", req["preset"])
			assert.Equal(t, "pipeline_123", req["pipeline_id"])

			props := req["properties"].(map[string]interface{})
			assert.Equal(t, "room-1", props["channel"])
			assert.Equal(t, "1", props["agent_rtc_uid"])

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","status":"RUNNING"}`))
			atomic.StoreInt32(&started, 1)
		case "/v2/projects/appid/agents/agent_123/turns":
			assert.Equal(t, int32(1), atomic.LoadInt32(&started))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"turns":[{"agent_id":"agent_123","turn_id":1}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)
	agoraClient := &AgoraClient{
		Agents:         rawClient.Agents,
		AppID:          "appid",
		AppCertificate: "app-cert",
		AuthMode:       AuthModeBasic,
	}

	agent := NewAgent(WithName("support-agent"))
	session := agent.CreateSession(agoraClient, CreateSessionOptions{
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
		Token:      "rtc-token",
		Preset: []string{
			AgentPresets.Asr.DeepgramNova3,
			AgentPresets.Llm.OpenAIGpt4oMini,
			AgentPresets.Tts.OpenAITts1,
		},
		PipelineID: "pipeline_123",
	})

	agentID, err := session.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "agent_123", agentID)

	turns, err := session.GetTurns(context.Background())
	require.NoError(t, err)
	require.Len(t, turns.Turns, 1)
	assert.Equal(t, "agent_123", *turns.Turns[0].AgentID)
}

func TestCreateSessionStartSendsManagedPresetPayloadWithoutGeneratedEmptyFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/projects/appid/join", r.URL.Path)
		var req map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "deepgram_nova_3,openai_gpt_4o_mini,minimax_speech_2_6_turbo", req["preset"])

		props := req["properties"].(map[string]interface{})
		llm := props["llm"].(map[string]interface{})
		tts := props["tts"].(map[string]interface{})
		asr := props["asr"].(map[string]interface{})
		assert.Equal(t, "openai", llm["style"])
		assert.Equal(t, "minimax", tts["vendor"])
		assert.Equal(t, "deepgram", asr["vendor"])
		payload, err := json.Marshal(req)
		require.NoError(t, err)
		assert.NotContains(t, string(payload), `"url":""`)
		assert.NotContains(t, string(payload), `"api_key":""`)
		assert.NotContains(t, string(payload), `"key":""`)
		assert.NotContains(t, string(payload), `"group_id":""`)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"agent_id":"agent_123","status":"RUNNING"}`))
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)
	agoraClient := &AgoraClient{
		Agents:         rawClient.Agents,
		AppID:          "appid",
		AppCertificate: "app-cert",
		AuthMode:       AuthModeBasic,
	}

	agent := NewAgent(WithName("managed-agent")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			Model:    "nova-3",
			Language: "en",
		})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			Model: "gpt-4o-mini",
		})).
		WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Model:   "speech_2_6_turbo",
			VoiceID: "English_captivating_female1",
		}))

	session := agent.CreateSession(agoraClient, CreateSessionOptions{
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
		Token:      "rtc-token",
	})

	agentID, err := session.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "agent_123", agentID)
}

func TestOffRemovesRegisteredHandler(t *testing.T) {
	session := NewAgentSession(AgentSessionOptions{
		Client:     nil,
		Agent:      NewAgent(),
		AppID:      "appid",
		Name:       "agent",
		Channel:    "room",
		AgentUID:   "1",
		RemoteUIDs: []string{"2"},
	})

	var count int
	handler := func(data interface{}) { count++ }
	session.On("started", handler)
	session.Off("started", handler)
	session.emit("started", map[string]string{"agent_id": "agent"})
	assert.Equal(t, 0, count)
}

func TestGeminiLiveMatchesTypeScriptShape(t *testing.T) {
	mllmTurnDetection := &Agora.StartAgentsRequestPropertiesMllmTurnDetection{
		Mode: Agora.StartAgentsRequestPropertiesMllmTurnDetectionModeServerVad.Ptr(),
		ServerVadConfig: &Agora.StartAgentsRequestPropertiesMllmTurnDetectionServerVadConfig{
			IdleTimeoutMs: Agora.Int(5000),
		},
	}
	config := vendors.NewGeminiLive(vendors.GeminiLiveOptions{
		APIKey:           "google-key",
		Model:            "gemini-live-2.5-flash",
		URL:              "wss://generativelanguage.googleapis.com/ws",
		Instructions:     "Be concise.",
		Voice:            "Aoede",
		GreetingMessage:  "Hello from Gemini",
		FailureMessage:   "Please try again.",
		InputModalities:  []string{"audio"},
		OutputModalities: []string{"text", "audio"},
		Messages: []map[string]interface{}{
			{"role": "system", "content": "short memory"},
		},
		AdditionalParams: map[string]interface{}{
			"temperature": 0.2,
		},
		TurnDetection: mllmTurnDetection,
	}).ToConfig()

	assert.Equal(t, map[string]interface{}{
		"vendor":  "gemini",
		"api_key": "google-key",
		"url":     "wss://generativelanguage.googleapis.com/ws",
		"params": map[string]interface{}{
			"temperature":  0.2,
			"model":        "gemini-live-2.5-flash",
			"instructions": "Be concise.",
			"voice":        "Aoede",
		},
		"messages": []map[string]interface{}{
			{"role": "system", "content": "short memory"},
		},
		"greeting_message":  "Hello from Gemini",
		"failure_message":   "Please try again.",
		"input_modalities":  []string{"audio"},
		"output_modalities": []string{"text", "audio"},
		"turn_detection":    mllmTurnDetection,
	}, config)
	assert.NotContains(t, config, "max_history")
	assert.NotContains(t, config, "predefined_tools")
}

func TestMLLMWrappersIncludeOptionalFields(t *testing.T) {
	openAIConfig := vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
		APIKey:         "key",
		URL:            "wss://openai.example.com/realtime",
		FailureMessage: "Retry",
	}).ToConfig()
	assert.Equal(t, "wss://openai.example.com/realtime", openAIConfig["url"])
	assert.Equal(t, "Retry", openAIConfig["failure_message"])
	assert.NotContains(t, openAIConfig, "style")
	assert.NotContains(t, openAIConfig, "predefined_tools")
	assert.NotContains(t, openAIConfig, "max_history")

	vertexConfig := vendors.NewVertexAI(vendors.VertexAIOptions{
		Model:               "gemini-live",
		URL:                 "wss://vertex.example.com/realtime",
		ProjectID:           "project",
		Location:            "us-central1",
		ADCredentialsString: "adc",
		FailureMessage:      "Try again",
	}).ToConfig()
	assert.Equal(t, "wss://vertex.example.com/realtime", vertexConfig["url"])
	assert.Equal(t, "Try again", vertexConfig["failure_message"])
	assert.NotContains(t, vertexConfig, "style")
	assert.NotContains(t, vertexConfig, "predefined_tools")
	assert.NotContains(t, vertexConfig, "max_history")
}

func TestXAIGrokMatchesV27Shape(t *testing.T) {
	sampleRate := 24000
	turnDetection := &Agora.StartAgentsRequestPropertiesMllmTurnDetection{
		Mode: Agora.StartAgentsRequestPropertiesMllmTurnDetectionModeServerVad.Ptr(),
	}
	config := vendors.NewXAIGrok(vendors.XAIGrokOptions{
		APIKey:          "xai-key",
		Voice:           "eve",
		Language:        "en",
		SampleRate:      &sampleRate,
		GreetingMessage: "hello",
		FailureMessage:  "try again",
		Params: map[string]interface{}{
			"temperature": 0.1,
		},
		TurnDetection: turnDetection,
	}).ToConfig()

	assert.Equal(t, "xai", config["vendor"])
	assert.Equal(t, "xai-key", config["api_key"])
	assert.Equal(t, "wss://api.x.ai/v1/realtime", config["url"])
	assert.Equal(t, map[string]interface{}{
		"temperature": 0.1,
		"voice":       "eve",
		"language":    "en",
		"sample_rate": 24000,
	}, config["params"])
	assert.Equal(t, "hello", config["greeting_message"])
	assert.Equal(t, "try again", config["failure_message"])
	assert.Equal(t, turnDetection, config["turn_detection"])
	assert.NotContains(t, config, "style")
	assert.NotContains(t, config, "predefined_tools")
	assert.NotContains(t, config, "max_history")
}

func TestThinkActionConstantsCoverV27Enums(t *testing.T) {
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnListeningActionInject, ThinkOnListeningActionInject)
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnListeningActionInterrupt, ThinkOnListeningActionInterrupt)
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnListeningActionIgnore, ThinkOnListeningActionIgnore)
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnThinkingActionInterrupt, ThinkOnThinkingActionInterrupt)
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnThinkingActionIgnore, ThinkOnThinkingActionIgnore)
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnSpeakingActionInterrupt, ThinkOnSpeakingActionInterrupt)
	assert.Equal(t, Agora.AgentThinkAgentManagementRequestOnSpeakingActionIgnore, ThinkOnSpeakingActionIgnore)
}

func TestInterruptionConstantsCoverV27Enums(t *testing.T) {
	assert.Equal(t, Agora.StartAgentsRequestPropertiesInterruptionModeStartOfSpeech, InterruptionModeStartOfSpeech)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesInterruptionModeKeywords, InterruptionModeKeywords)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesInterruptionDisabledConfigStrategyAppend, InterruptionDisabledStrategyAppend)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesInterruptionDisabledConfigStrategyIgnore, InterruptionDisabledStrategyIgnore)
}

func TestSpeakPriorityConstantsCoverV27Enums(t *testing.T) {
	assert.Equal(t, Agora.SpeakAgentsRequestPriorityInterrupt, SpeakPriorityInterrupt)
	assert.Equal(t, Agora.SpeakAgentsRequestPriorityAppend, SpeakPriorityAppend)
	assert.Equal(t, Agora.SpeakAgentsRequestPriorityIgnore, SpeakPriorityIgnore)
}

func TestMllmTurnDetectionModeConstantsCoverV27Enums(t *testing.T) {
	assert.Equal(t, Agora.StartAgentsRequestPropertiesMllmTurnDetectionModeAgoraVad, MllmTurnDetectionModeAgoraVad)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesMllmTurnDetectionModeServerVad, MllmTurnDetectionModeServerVad)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesMllmTurnDetectionModeSemanticVad, MllmTurnDetectionModeSemanticVad)
}

func TestAzureOpenAIIncludesModelInParams(t *testing.T) {
	llm := vendors.NewAzureOpenAI(vendors.AzureOpenAIOptions{
		APIKey:         "azure-key",
		Endpoint:       "https://example.openai.azure.com",
		DeploymentName: "deploy-1",
		APIVersion:     "2024-08-01-preview",
		Model:          "gpt-4o",
	})
	cfg := llm.ToConfig()
	params, _ := cfg["params"].(map[string]interface{})
	require.NotNil(t, params)
	assert.Equal(t, "gpt-4o", params["model"], "AzureOpenAI must emit params.model when Model is set, matching TS parity")
}

func TestAzureOpenAIParamsOverrideModel(t *testing.T) {
	llm := vendors.NewAzureOpenAI(vendors.AzureOpenAIOptions{
		APIKey:         "azure-key",
		Endpoint:       "https://example.openai.azure.com",
		DeploymentName: "deploy-1",
		APIVersion:     "2024-08-01-preview",
		Model:          "gpt-4o",
		Params:         map[string]interface{}{"model": "gpt-4o-mini"},
	})
	cfg := llm.ToConfig()
	params, _ := cfg["params"].(map[string]interface{})
	require.NotNil(t, params)
	assert.Equal(t, "gpt-4o-mini", params["model"], "Explicit Params['model'] must override named Model")
}

func TestOpenAIRealtimeUserParamsOverrideModel(t *testing.T) {
	o := vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
		APIKey: "openai-key",
		Model:  "gpt-4o-realtime-preview",
		Params: map[string]interface{}{"model": "custom-realtime"},
	})
	cfg := o.ToConfig()
	params, _ := cfg["params"].(map[string]interface{})
	require.NotNil(t, params)
	assert.Equal(t, "custom-realtime", params["model"], "explicit Params['model'] must override named Model, matching TS XaiGrok/OpenAIRealtime behavior")
}

func TestAvatarAdditionalParamsCannotOverwriteRequiredFields(t *testing.T) {
	heygen := vendors.NewHeyGenAvatar(vendors.HeyGenAvatarOptions{
		APIKey:   "heygen-key",
		Quality:  "high",
		AgoraUID: "100",
		AdditionalParams: map[string]interface{}{
			"api_key":   "evil",
			"quality":   "broken",
			"agora_uid": "999",
		},
	})
	hParams, _ := heygen.ToConfig()["params"].(map[string]interface{})
	require.NotNil(t, hParams)
	assert.Equal(t, "heygen-key", hParams["api_key"])
	assert.Equal(t, "high", hParams["quality"])
	assert.Equal(t, "100", hParams["agora_uid"])

	live := vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
		APIKey:   "live-key",
		Quality:  "medium",
		AgoraUID: "200",
		AdditionalParams: map[string]interface{}{
			"api_key":   "evil",
			"quality":   "broken",
			"agora_uid": "999",
		},
	})
	lParams, _ := live.ToConfig()["params"].(map[string]interface{})
	require.NotNil(t, lParams)
	assert.Equal(t, "live-key", lParams["api_key"])
	assert.Equal(t, "medium", lParams["quality"])
	assert.Equal(t, "200", lParams["agora_uid"])

	akool := vendors.NewAkoolAvatar(vendors.AkoolAvatarOptions{
		APIKey: "akool-key",
		AdditionalParams: map[string]interface{}{
			"api_key": "evil",
		},
	})
	aParams, _ := akool.ToConfig()["params"].(map[string]interface{})
	require.NotNil(t, aParams)
	assert.Equal(t, "akool-key", aParams["api_key"])

	anam := vendors.NewAnamAvatar(vendors.AnamAvatarOptions{
		APIKey: "anam-key",
		AdditionalParams: map[string]interface{}{
			"api_key": "evil",
		},
	})
	nParams, _ := anam.ToConfig()["params"].(map[string]interface{})
	require.NotNil(t, nParams)
	assert.Equal(t, "anam-key", nParams["api_key"])

	generic := vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
		APIKey:     "generic-key",
		APIBaseURL: "https://avatar.example.com",
		AvatarID:   "generic-avatar",
		AgoraUID:   "200",
		AdditionalParams: map[string]interface{}{
			"api_key":      "evil",
			"api_base_url": "https://evil.example.com",
			"avatar_id":    "evil-avatar",
			"agora_uid":    "999",
		},
	})
	gParams, _ := generic.ToConfig()["params"].(map[string]interface{})
	require.NotNil(t, gParams)
	assert.Equal(t, "generic-key", gParams["api_key"])
	assert.Equal(t, "https://avatar.example.com", gParams["api_base_url"])
	assert.Equal(t, "generic-avatar", gParams["avatar_id"])
	assert.Equal(t, "200", gParams["agora_uid"])
}

func TestAgoraClientExposesAllGeneratedSubClients(t *testing.T) {
	c := NewAgoraClient(AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          "0123456789abcdef0123456789abcdef",
		AppCertificate: "appcert",
	})
	require.NotNil(t, c)
	assert.NotNil(t, c.Agents)
	assert.NotNil(t, c.AgentManagement)
	assert.NotNil(t, c.Telephony, "AgoraClient must expose Telephony for v2.7 outbound calls")
	assert.NotNil(t, c.PhoneNumbers, "AgoraClient must expose PhoneNumbers for v2.7 number management")
}

func TestWithInterruptionForwardsConfig(t *testing.T) {
	interruption := &InterruptionConfig{
		Enable: Agora.Bool(false),
		DisabledConfig: &Agora.StartAgentsRequestPropertiesInterruptionDisabledConfig{
			Strategy: Agora.StartAgentsRequestPropertiesInterruptionDisabledConfigStrategyIgnore.Ptr(),
		},
	}

	props, err := NewAgent(WithInterruptionConfig(interruption)).ToProperties(ToPropertiesOptions{
		Channel:              "room",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)
	require.NotNil(t, props.Interruption)
	require.NotNil(t, props.Interruption.Enable)
	assert.False(t, *props.Interruption.Enable)
	require.NotNil(t, props.Interruption.DisabledConfig)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesInterruptionDisabledConfigStrategyIgnore, *props.Interruption.DisabledConfig.Strategy)

	props, err = NewAgent().WithInterruption(interruption).ToProperties(ToPropertiesOptions{
		Channel:              "room",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)
	require.NotNil(t, props.Interruption)
	require.NotNil(t, props.Interruption.Enable)
	assert.False(t, *props.Interruption.Enable)
	require.NotNil(t, props.Interruption.DisabledConfig)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesInterruptionDisabledConfigStrategyIgnore, *props.Interruption.DisabledConfig.Strategy)
}

func TestPresetBackedOpenAIVendorsAllowMissingKeys(t *testing.T) {
	agent := NewAgent(WithInstructions("Preset-backed flow")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
			Model: "nova-3",
		})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			Model: "gpt-5-mini",
		})).
		WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
			Voice: "alloy",
		}))

	props, err := agent.ToProperties(ToPropertiesOptions{
		Channel:              "room-1",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	preset, resolved, err := ResolveSessionPresets(nil, props)
	require.NoError(t, err)
	assert.Equal(t, "deepgram_nova_3,openai_gpt_5_mini,openai_tts_1", preset)
	require.NotNil(t, resolved)

	payload, err := json.Marshal(resolved)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "api_key")
}

func TestPresetBackedMiniMaxTTSAllowsMissingKey(t *testing.T) {
	tts := vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
		Model:   "speech-2.6-turbo",
		VoiceID: "English_captivating_female1",
		URL:     "wss://api-uw.minimax.io/ws/v1/t2a_v2",
	}).ToConfig()

	assert.Equal(t, "minimax", tts["vendor"])
	params := tts["params"].(map[string]interface{})
	assert.Equal(t, "speech-2.6-turbo", params["model"])
	assert.NotContains(t, params, "key")
	assert.NotContains(t, params, "group_id")
	assert.Equal(t, map[string]interface{}{"voice_id": "English_captivating_female1"}, params["voice_setting"])
	assert.Equal(t, "wss://api-uw.minimax.io/ws/v1/t2a_v2", params["url"])
}

func TestManagedPresetPayloadOmitsProviderOwnedFields(t *testing.T) {
	smartFormat := true
	punctuation := true
	maxHistory := 15
	agent := NewAgent(
		WithInstructions("Preset flow"),
		WithGreeting("hello"),
	).WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
		Model:       "nova-3",
		Language:    "en",
		SmartFormat: &smartFormat,
		Punctuation: &punctuation,
	})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		Model:      "gpt-4o-mini",
		MaxHistory: &maxHistory,
		Params: map[string]interface{}{
			"max_tokens":  1024,
			"temperature": 0.7,
			"top_p":       0.95,
		},
	})).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
		Model:   "speech_2_6_turbo",
		VoiceID: "English_captivating_female1",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	preset, resolved, err := ResolveSessionPresetsMap(nil, props)
	require.NoError(t, err)
	assert.Equal(t, "deepgram_nova_3,openai_gpt_4o_mini,minimax_speech_2_6_turbo", preset)

	payload, err := json.Marshal(resolved)
	require.NoError(t, err)
	payloadText := string(payload)
	assert.NotContains(t, payloadText, `"url":""`)
	assert.NotContains(t, payloadText, `"api_key":""`)
	assert.NotContains(t, payloadText, `"key":""`)
	assert.NotContains(t, payloadText, `"group_id":""`)
	assert.NotContains(t, payloadText, `"model":"gpt-4o-mini"`)
	assert.NotContains(t, payloadText, `"model":"speech_2_6_turbo"`)
	assert.NotContains(t, payloadText, `"model":"nova-3"`)

	asr := resolved["asr"].(map[string]interface{})
	asrParams := asr["params"].(map[string]interface{})
	assert.Equal(t, "deepgram", asr["vendor"])
	assert.Equal(t, "en", asrParams["language"])
	assert.Equal(t, true, asrParams["smart_format"])
	assert.Equal(t, true, asrParams["punctuation"])
	assert.NotContains(t, asrParams, "api_key")
	assert.NotContains(t, asrParams, "model")

	llm := resolved["llm"].(map[string]interface{})
	llmParams := llm["params"].(map[string]interface{})
	assert.Equal(t, "openai", llm["style"])
	assert.Equal(t, 1024, llmParams["max_tokens"])
	assert.Equal(t, 0.7, llmParams["temperature"])
	assert.Equal(t, 0.95, llmParams["top_p"])
	assert.Equal(t, 15, llm["max_history"])
	assert.NotContains(t, llm, "url")
	assert.NotContains(t, llm, "api_key")
	assert.NotContains(t, llmParams, "model")

	tts := resolved["tts"].(map[string]interface{})
	ttsParams := tts["params"].(map[string]interface{})
	assert.Equal(t, "minimax", tts["vendor"])
	assert.Equal(t, map[string]interface{}{"voice_id": "English_captivating_female1"}, ttsParams["voice_setting"])
	assert.NotContains(t, ttsParams, "key")
	assert.NotContains(t, ttsParams, "group_id")
	assert.NotContains(t, ttsParams, "url")
	assert.NotContains(t, ttsParams, "model")
}

func TestManagedOpenAITTSOmitKeyAndModel(t *testing.T) {
	agent := NewAgent().WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
		Model: "nova-2",
	})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		Model: "gpt-5-mini",
	})).WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
		Model: "tts-1",
		Voice: "alloy",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	preset, resolved, err := ResolveSessionPresetsMap(nil, props)
	require.NoError(t, err)
	assert.Equal(t, "deepgram_nova_2,openai_gpt_5_mini,openai_tts_1", preset)

	tts := resolved["tts"].(map[string]interface{})
	params := tts["params"].(map[string]interface{})
	assert.Equal(t, "alloy", params["voice"])
	assert.NotContains(t, params, "key")
	assert.NotContains(t, params, "api_key")
	assert.NotContains(t, params, "model")
}

func TestDeepgramTTSVendorConfig(t *testing.T) {
	sampleRate := vendors.SampleRate24kHz
	tts := vendors.NewDeepgramTTS(vendors.DeepgramTTSOptions{
		APIKey:     "deepgram-key",
		Model:      "aura-2-thalia-en",
		BaseURL:    "wss://api.deepgram.com/v1/speak",
		SampleRate: &sampleRate,
		Params: map[string]interface{}{
			"encoding": "linear16",
		},
	}).ToConfig()

	assert.Equal(t, "deepgram", tts["vendor"])
	assert.Equal(t, map[string]interface{}{
		"api_key":     "deepgram-key",
		"model":       "aura-2-thalia-en",
		"base_url":    "wss://api.deepgram.com/v1/speak",
		"sample_rate": 24000,
		"encoding":    "linear16",
	}, tts["params"])
}

func TestAresASRRemainsKeylessWithoutPreset(t *testing.T) {
	agent := NewAgent().WithStt(vendors.NewAresSTT(vendors.AresSTTOptions{
		Language: "en-US",
		AdditionalParams: map[string]interface{}{
			"sample_rate": 16000,
		},
	})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		Model: "gpt-4o-mini",
	})).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
		Model:   "speech_2_8_turbo",
		VoiceID: "English_captivating_female1",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "rtc-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	preset, resolved, err := ResolveSessionPresetsMap(nil, props)
	require.NoError(t, err)
	assert.Equal(t, "openai_gpt_4o_mini,minimax_speech_2_8_turbo", preset)

	asr := resolved["asr"].(map[string]interface{})
	params := asr["params"].(map[string]interface{})
	assert.Equal(t, "ares", asr["vendor"])
	assert.Equal(t, "en-US", asr["language"])
	assert.Equal(t, 16000, params["sample_rate"])
	assert.NotContains(t, params, "api_key")
	assert.NotContains(t, params, "key")
	assert.NotContains(t, params, "model")
}

func TestBYOKProvidersAreNotTreatedAsManagedPresets(t *testing.T) {
	agent := NewAgent().WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
		APIKey: "deepgram-key",
		Model:  "nova-3",
	})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		APIKey:  "openai-key",
		Model:   "gpt-4o-mini",
		Headers: map[string]string{"X-Trace-Id": "trace-123"},
	})).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
		Key:     "minimax-key",
		GroupID: "minimax-group",
		Model:   "speech_2_6_turbo",
		VoiceID: "English_captivating_female1",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:    "room-1",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)

	preset, resolved, err := ResolveSessionPresetsMap(nil, props)
	require.NoError(t, err)
	assert.Empty(t, preset)

	asrParams := resolved["asr"].(map[string]interface{})["params"].(map[string]interface{})
	llm := resolved["llm"].(map[string]interface{})
	ttsParams := resolved["tts"].(map[string]interface{})["params"].(map[string]interface{})
	assert.Equal(t, "deepgram-key", asrParams["api_key"])
	assert.Equal(t, "openai-key", llm["api_key"])
	assert.Equal(t, map[string]string{"X-Trace-Id": "trace-123"}, llm["headers"])
	assert.Equal(t, "minimax-key", ttsParams["key"])
	assert.Equal(t, "minimax-group", ttsParams["group_id"])
}

func TestWithMllmSetsMllmEnableWithoutLegacyFlag(t *testing.T) {
	props, err := NewAgent().WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
		APIKey: "openai-key",
	})).ToProperties(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)
	require.NotNil(t, props.Mllm)
	require.NotNil(t, props.Mllm.Enable)
	assert.True(t, *props.Mllm.Enable)
	assert.Nil(t, props.AdvancedFeatures)
}

func TestWithMllmForcesEnableAndRemovesDeprecatedAdvancedFlag(t *testing.T) {
	enableMllm := true
	enableRtm := true
	props, err := NewAgent(WithAdvancedFeatures(&AdvancedFeatures{
		EnableMllm: &enableMllm,
		EnableRtm:  &enableRtm,
	})).WithMllm(disabledMllmVendor{}).ToProperties(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)
	require.NotNil(t, props.Mllm)
	require.NotNil(t, props.Mllm.Enable)
	assert.True(t, *props.Mllm.Enable)
	require.NotNil(t, props.AdvancedFeatures)
	assert.Nil(t, props.AdvancedFeatures.EnableMllm)
	require.NotNil(t, props.AdvancedFeatures.EnableRtm)
	assert.True(t, *props.AdvancedFeatures.EnableRtm)
	require.NotNil(t, props.Parameters)
	require.NotNil(t, props.Parameters.DataChannel)
	assert.Equal(t, Agora.StartAgentsRequestPropertiesParametersDataChannelRtm, *props.Parameters.DataChannel)
}

func TestWithMllmDropsAdvancedFeaturesWhenOnlyDeprecatedEnableMllmWasSet(t *testing.T) {
	enableMllm := true
	props, err := NewAgent(WithAdvancedFeatures(&AdvancedFeatures{
		EnableMllm: &enableMllm,
	})).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
		APIKey: "openai-key",
	})).ToProperties(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)
	require.NotNil(t, props.Mllm)
	require.NotNil(t, props.Mllm.Enable)
	assert.True(t, *props.Mllm.Enable)
	assert.Nil(t, props.AdvancedFeatures)
}

func TestMllmModeDoesNotRequireLlmOrTtsWhenEnableMissing(t *testing.T) {
	agent := NewAgent()
	agent.mllm = map[string]interface{}{
		"vendor": "openai",
	}

	props, err := agent.ToProperties(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)
	require.NotNil(t, props.Mllm)
}

func TestMllmWithEnabledAvatarIsRejectedWithoutRequiringTts(t *testing.T) {
	agent := NewAgent().
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey: "openai-key",
		})).
		WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
			APIKey:   "live-key",
			Quality:  "high",
			AgoraUID: "42",
		}))

	_, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "avatar is only supported with cascading")
	assert.NotContains(t, err.Error(), "TTS configuration is required")
}

func TestMllmWithDisabledAvatarDoesNotRequireTts(t *testing.T) {
	enable := false
	agent := NewAgent().
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey: "openai-key",
		})).
		WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
			APIKey:   "live-key",
			Quality:  "high",
			AgoraUID: "42",
			Enable:   &enable,
		}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)
	assert.Contains(t, props, "mllm")
	assert.Contains(t, props, "avatar")
}

func TestToPropertiesBubblesMLLMFieldsAndPreservesVendorOverrides(t *testing.T) {
	maxHistory := 9
	agent := NewAgent(
		WithGreeting("Agent greeting"),
		WithFailureMessage("Agent failure"),
		WithMaxHistory(maxHistory),
	).WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
		APIKey:          "openai-key",
		Model:           "gpt-4o-realtime-preview",
		URL:             "wss://openai.example.com/realtime",
		GreetingMessage: "Vendor greeting",
	}))

	props, err := agent.ToProperties(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)
	require.NotNil(t, props)
	require.NotNil(t, props.Mllm)

	payload, err := json.Marshal(props.Mllm)
	require.NoError(t, err)
	assert.Contains(t, string(payload), "greeting_message")
	assert.Contains(t, string(payload), "url")

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, "Vendor greeting", decoded["greeting_message"])
	assert.Equal(t, "wss://openai.example.com/realtime", decoded["url"])
	assert.NotContains(t, decoded, "max_history")
}

func TestToPropertiesLlmAgentFieldsOverrideVendorDefaults(t *testing.T) {
	maxHistory := 9
	vendorMaxHistory := 3
	props, err := NewAgent(
		WithGreeting("Agent greeting"),
		WithFailureMessage("Agent failure"),
		WithMaxHistory(maxHistory),
	).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		APIKey:          "openai-key",
		Model:           "gpt-4o-mini",
		GreetingMessage: "Vendor greeting",
		FailureMessage:  "Vendor failure",
		MaxHistory:      &vendorMaxHistory,
	})).WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
		APIKey: "openai-key",
		Voice:  "alloy",
	})).ToPropertiesMap(ToPropertiesOptions{
		Channel:    "room",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)

	llm := props["llm"].(map[string]interface{})
	assert.Equal(t, "Agent greeting", llm["greeting_message"])
	assert.Equal(t, "Agent failure", llm["failure_message"])
	assert.Equal(t, 9, llm["max_history"])
}

func TestAvatarHelpersCoverLiveAvatarAndAnam(t *testing.T) {
	assert.True(t, IsLiveAvatarAvatar("liveavatar"))
	assert.True(t, IsAnamAvatar("anam"))
	assert.True(t, IsGenericAvatar("generic"))
	require.NoError(t, ValidateAvatarConfig("liveavatar", map[string]interface{}{
		"api_key":   "live-key",
		"quality":   "high",
		"agora_uid": float64(42),
	}))
	require.NoError(t, ValidateAvatarConfig("anam", map[string]interface{}{
		"api_key": "anam-key",
	}))
	require.Error(t, ValidateAvatarConfig("generic", map[string]interface{}{
		"api_key":      "generic-key",
		"api_base_url": "https://avatar.example.com",
		"avatar_id":    "",
		"agora_uid":    float64(42),
	}))
	require.NoError(t, ValidateAvatarConfig("generic", map[string]interface{}{
		"api_key":      "generic-key",
		"api_base_url": "https://avatar.example.com",
		"avatar_id":    "avatar-1",
		"agora_uid":    float64(42),
	}))
	require.NoError(t, ValidateTtsSampleRate("liveavatar", 24000))
	require.Error(t, ValidateTtsSampleRate("liveavatar", 16000))

	avatar := vendors.NewAnamAvatar(vendors.AnamAvatarOptions{
		APIKey:    "anam-key",
		PersonaID: "persona-1",
	}).ToConfig()
	assert.Equal(t, "anam", avatar["vendor"])

	generic := vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
		APIKey:     "generic-key",
		APIBaseURL: "https://avatar.example.com",
		AvatarID:   "avatar-1",
		AgoraUID:   "42",
		AdditionalParams: map[string]interface{}{
			"avatar_id": "should-not-win",
			"custom":    "value",
		},
	}).ToConfig()
	assert.Equal(t, "generic", generic["vendor"])
	genericParams := generic["params"].(map[string]interface{})
	assert.Equal(t, "generic-key", genericParams["api_key"])
	assert.Equal(t, "https://avatar.example.com", genericParams["api_base_url"])
	assert.Equal(t, "avatar-1", genericParams["avatar_id"])
	assert.Equal(t, "42", genericParams["agora_uid"])
	assert.Equal(t, "value", genericParams["custom"])
	assert.NotContains(t, genericParams, "agora_appid")
	assert.NotContains(t, genericParams, "agora_channel")
	assert.NotContains(t, genericParams, "agora_token")
	assert.Zero(t, vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
		APIKey:     "generic-key",
		APIBaseURL: "https://avatar.example.com",
		AvatarID:   "avatar-1",
		AgoraUID:   "42",
	}).RequiredSampleRate())
	require.NoError(t, ValidateTtsSampleRate("generic", 16000))
	require.NoError(t, ValidateTtsSampleRate("generic", 24000))
}

func TestGenericAvatarDoesNotEnforceSampleRate(t *testing.T) {
	sampleRate := vendors.SampleRate16kHz
	require.NotPanics(t, func() {
		_ = NewAgent().
			WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
				Key:        "eleven-key",
				ModelID:    "eleven_flash_v2_5",
				VoiceID:    "voice-id",
				SampleRate: &sampleRate,
			})).
			WithAvatar(vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
				APIKey:     "generic-key",
				APIBaseURL: "https://avatar.example.com",
				AvatarID:   "avatar-1",
				AgoraUID:   "42",
			}))
	})
}

func TestDisabledAvatarSkipsSampleRateValidationAndEnrichment(t *testing.T) {
	enable := false
	sampleRate := vendors.SampleRate16kHz
	agent := NewAgent().
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:        "eleven-key",
			ModelID:    "eleven_flash_v2_5",
			VoiceID:    "voice-id",
			SampleRate: &sampleRate,
		})).
		WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
			APIKey:  "live-key",
			Quality: "high",
			// LiveAvatar normally requires an avatar UID and 24 kHz TTS, but a disabled
			// avatar should be passed through without session-time enrichment.
			AgoraUID: "42",
			Enable:   &enable,
		}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	params := props["avatar"].(map[string]interface{})["params"].(map[string]interface{})
	assert.NotContains(t, params, "agora_token")
}

func TestGenericAvatarEnrichmentAddsAppChannelAndConvoAIToken(t *testing.T) {
	agent := NewAgent().WithAvatar(vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
		APIKey:     "generic-key",
		APIBaseURL: "https://avatar.example.com",
		AvatarID:   "avatar-1",
		AgoraUID:   "42",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	avatar := props["avatar"].(map[string]interface{})
	params := avatar["params"].(map[string]interface{})
	assert.Equal(t, "0123456789abcdef0123456789abcdef", params["agora_appid"])
	assert.Equal(t, "room-1", params["agora_channel"])
	assert.NotEmpty(t, params["agora_token"])
	assert.NotEqual(t, props["token"], params["agora_token"])
}

func TestAvatarEnrichmentSupportsNumericAgoraUID(t *testing.T) {
	agent := NewAgent()
	agent.avatar = map[string]interface{}{
		"enable": true,
		"vendor": "generic",
		"params": map[string]interface{}{
			"api_key":      "generic-key",
			"api_base_url": "https://avatar.example.com",
			"avatar_id":    "avatar-1",
			"agora_uid":    float64(42),
		},
	}

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	params := props["avatar"].(map[string]interface{})["params"].(map[string]interface{})
	assert.NotEmpty(t, params["agora_token"])
}

func TestToPropertiesDirectPathUsesAvatarEnrichment(t *testing.T) {
	agent := NewAgent().WithAvatar(vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
		APIKey:     "generic-key",
		APIBaseURL: "https://avatar.example.com",
		AvatarID:   "avatar-1",
		AgoraUID:   "42",
	}))

	props, err := agent.ToProperties(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		SkipVendorValidation: true,
	})
	require.NoError(t, err)
	require.NotNil(t, props.Avatar)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", props.Avatar.Params["agora_appid"])
	assert.Equal(t, "room-1", props.Avatar.Params["agora_channel"])
	assert.NotEmpty(t, props.Avatar.Params["agora_token"])
}

func TestValidateEnrichedAvatarConfigRequiresGeneratedFields(t *testing.T) {
	err := validateEnrichedAvatarConfig(map[string]interface{}{
		"avatar": map[string]interface{}{
			"enable": true,
			"vendor": "generic",
			"params": map[string]interface{}{
				"api_key":      "generic-key",
				"api_base_url": "https://avatar.example.com",
				"avatar_id":    "avatar-1",
				"agora_uid":    "42",
			},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agora_appid")
}

func TestAppCredentialsModeRequiresAppCertificate(t *testing.T) {
	session := NewAgentSession(AgentSessionOptions{
		Agent:                    NewAgent(),
		AppID:                    "appid",
		Name:                     "agent",
		Channel:                  "room-1",
		AgentUID:                 "1",
		RemoteUIDs:               []string{"2"},
		UseAppCredentialsForREST: true,
	})

	_, err := session.convoAIRequestOpts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "appCertificate is required")
}

func TestAvatarEnrichmentExplainsMissingAppCertificate(t *testing.T) {
	agent := NewAgent().WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
		APIKey:   "live-key",
		Quality:  "high",
		AgoraUID: "42",
	}))

	_, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		SkipVendorValidation: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot auto-generate avatar agora_token")
}

func TestAvatarEnrichmentValidatesExpiresInEvenWithSessionToken(t *testing.T) {
	agent := NewAgent().WithAvatar(vendors.NewGenericAvatar(vendors.GenericAvatarOptions{
		APIKey:     "generic-key",
		APIBaseURL: "https://avatar.example.com",
		AvatarID:   "avatar-1",
		AgoraUID:   "42",
	}))

	_, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		ExpiresIn:            -1,
		SkipVendorValidation: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expiresIn")
}

func TestAvatarEnrichmentPreservesProvidedTokenAndWarnsOnUidCollision(t *testing.T) {
	var warnings []string
	agent := NewAgent().WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
		APIKey:     "live-key",
		Quality:    "high",
		AgoraUID:   "1",
		AgoraToken: "avatar-token",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		SkipVendorValidation: true,
		Warn: func(msg string) {
			warnings = append(warnings, msg)
		},
	})
	require.NoError(t, err)

	params := props["avatar"].(map[string]interface{})["params"].(map[string]interface{})
	assert.Equal(t, "avatar-token", params["agora_token"])
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "avatar agora_uid matches agent_rtc_uid")
}

func TestLiveAvatarEnrichmentGeneratesTokenOnly(t *testing.T) {
	agent := NewAgent().WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
		APIKey:   "live-key",
		Quality:  "high",
		AgoraUID: "42",
	}))

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:              "room-1",
		Token:                "agent-token",
		AgentUID:             "1",
		RemoteUIDs:           []string{"100"},
		AppID:                "0123456789abcdef0123456789abcdef",
		AppCertificate:       "fedcba9876543210fedcba9876543210",
		SkipVendorValidation: true,
	})
	require.NoError(t, err)

	params := props["avatar"].(map[string]interface{})["params"].(map[string]interface{})
	assert.NotEmpty(t, params["agora_token"])
	assert.NotContains(t, params, "agora_appid")
	assert.NotContains(t, params, "agora_channel")
}

func TestSessionWarnsForAvatarWithoutExplicitSampleRateAndSupportsWarnHook(t *testing.T) {
	var warnings []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/projects/0123456789abcdef0123456789abcdef/join":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","status":"RUNNING"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)

	agent := NewAgent(WithName("avatar-agent")).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey: "openai-key",
			Model:  "gpt-4o-mini",
		})).
		WithTts(vendors.NewMicrosoftTTS(vendors.MicrosoftTTSOptions{
			Key:       "ms-key",
			Region:    "eastus",
			VoiceName: "en-US-JennyNeural",
		})).
		WithAvatar(vendors.NewLiveAvatarAvatar(vendors.LiveAvatarAvatarOptions{
			APIKey:   "live-key",
			Quality:  "high",
			AgoraUID: "42",
		}))

	session := NewAgentSession(AgentSessionOptions{
		Client:         rawClient.Agents,
		Agent:          agent,
		AppID:          "0123456789abcdef0123456789abcdef",
		AppCertificate: "fedcba9876543210fedcba9876543210",
		Name:           "avatar-agent",
		Channel:        "room",
		Token:          "rtc-token",
		AgentUID:       "1",
		RemoteUIDs:     []string{"2"},
		Warn: func(msg string) {
			warnings = append(warnings, msg)
		},
	})

	_, err := session.Start(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, warnings)
	assert.Contains(t, warnings[0], "LiveAvatar")
}

func TestSessionWarnHookReceivesHandlerPanics(t *testing.T) {
	var warnings []string

	session := NewAgentSession(AgentSessionOptions{
		Client:     nil,
		Agent:      NewAgent(),
		AppID:      "appid",
		Name:       "agent",
		Channel:    "room",
		AgentUID:   "1",
		RemoteUIDs: []string{"2"},
		Warn: func(msg string) {
			warnings = append(warnings, msg)
		},
	})

	session.On("started", func(data interface{}) {
		panic("boom")
	})
	session.emit("started", map[string]string{"agent_id": "agent"})
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "recovered panic")
}

func TestSessionThinkRoutesToAgentManagement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/projects/appid/agents/agent_123/think":
			var req map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "Injected instruction", req["text"])
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","channel":"room-1","start_ts":123}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)

	session := NewAgentSession(AgentSessionOptions{
		Client:                rawClient.Agents,
		AgentManagementClient: rawClient.AgentManagement,
		Agent:                 NewAgent(),
		AppID:                 "appid",
		Name:                  "agent",
		Channel:               "room-1",
		AgentUID:              "1",
		RemoteUIDs:            []string{"2"},
	})
	session.status = StatusRunning
	session.agentID = "agent_123"

	resp, err := session.Think(context.Background(), "Injected instruction", nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "agent_123", *resp.AgentID)
}

func TestSessionThinkWithOptionsForwardsFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/projects/appid/agents/agent_123/think":
			var req map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "Injected instruction", req["text"])
			assert.Equal(t, "interrupt", req["on_thinking_action"])
			assert.Equal(t, false, req["interruptable"])
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","channel":"room-1","start_ts":123}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)

	session := NewAgentSession(AgentSessionOptions{
		Client:                rawClient.Agents,
		AgentManagementClient: rawClient.AgentManagement,
		Agent:                 NewAgent(),
		AppID:                 "appid",
		Name:                  "agent",
		Channel:               "room-1",
		AgentUID:              "1",
		RemoteUIDs:            []string{"2"},
	})
	session.status = StatusRunning
	session.agentID = "agent_123"

	onThinking := Agora.AgentThinkAgentManagementRequestOnThinkingActionInterrupt
	notInterruptable := false
	resp, err := session.ThinkWithOptions(context.Background(), "Injected instruction", &ThinkOptions{
		OnThinkingAction: &onThinking,
		Interruptable:    &notInterruptable,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "agent_123", *resp.AgentID)
}

func TestGetTurnsForwardsPaginationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/projects/appid/agents/agent_123/turns", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page_index"))
		assert.Equal(t, "25", r.URL.Query().Get("page_size"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"agent_id":"agent_123","turns":[],"pagination":{"page_index":2,"total_pages":2,"is_last_page":true}}`))
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)
	session := NewAgentSession(AgentSessionOptions{
		Client:     rawClient.Agents,
		Agent:      NewAgent(),
		AppID:      "appid",
		Name:       "agent",
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"2"},
	})
	session.agentID = "agent_123"

	pageIndex := 2
	pageSize := 25
	resp, err := session.GetTurns(context.Background(), GetTurnsOptions{
		PageIndex: &pageIndex,
		PageSize:  &pageSize,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Pagination)
	assert.True(t, *resp.Pagination.IsLastPage)
}

func TestGetAllTurnsAggregatesPages(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/projects/appid/agents/agent_123/turns", r.URL.Path)
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page_index") {
		case "1":
			assert.Equal(t, "50", r.URL.Query().Get("page_size"))
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","turns":[{"agent_id":"agent_123","turn_id":1}],"pagination":{"page_index":1,"total_pages":2,"is_last_page":false}}`))
		case "2":
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","turns":[{"agent_id":"agent_123","turn_id":2}],"pagination":{"page_index":2,"total_pages":2,"is_last_page":true}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)
	session := NewAgentSession(AgentSessionOptions{
		Client:     rawClient.Agents,
		Agent:      NewAgent(),
		AppID:      "appid",
		Name:       "agent",
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"2"},
	})
	session.agentID = "agent_123"

	resp, err := session.GetAllTurns(context.Background())
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Turns, 2)
	require.NotNil(t, resp.Pagination)
	assert.True(t, *resp.Pagination.IsLastPage)
	assert.Equal(t, 2, calls)
}

func TestGetAllTurnsStopsWhenPaginationMissing(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/projects/appid/agents/agent_123/turns", r.URL.Path)
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"agent_id":"agent_123","turns":[{"agent_id":"agent_123","turn_id":1}]}`))
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)
	session := NewAgentSession(AgentSessionOptions{
		Client:     rawClient.Agents,
		Agent:      NewAgent(),
		AppID:      "appid",
		Name:       "agent",
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"2"},
	})
	session.agentID = "agent_123"

	resp, err := session.GetAllTurns(context.Background())
	require.NoError(t, err)
	require.Len(t, resp.Turns, 1)
	assert.Equal(t, 1, calls)
}

func TestGetAllTurnsErrorsWhenPaginationDoesNotAdvance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/projects/appid/agents/agent_123/turns", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page_index") {
		case "1":
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","turns":[{"agent_id":"agent_123","turn_id":1}],"pagination":{"page_index":1,"is_last_page":false}}`))
		case "2":
			_, _ = w.Write([]byte(`{"agent_id":"agent_123","turns":[{"agent_id":"agent_123","turn_id":2}],"pagination":{"page_index":1,"is_last_page":false}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	rawClient := client.NewClient(
		option.WithBaseURL(server.URL),
		option.WithBasicAuth("user", "pass"),
		option.WithMaxAttempts(1),
	)
	session := NewAgentSession(AgentSessionOptions{
		Client:     rawClient.Agents,
		Agent:      NewAgent(),
		AppID:      "appid",
		Name:       "agent",
		Channel:    "room-1",
		AgentUID:   "1",
		RemoteUIDs: []string{"2"},
	})
	session.agentID = "agent_123"

	_, err := session.GetAllTurns(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pagination did not advance")
}

func TestWithGreetingConfigsAndPauseStateEnabledSerializeToMap(t *testing.T) {
	interruptable := true
	mode := Agora.StartAgentsRequestPropertiesLlmGreetingConfigsModeSingleEvery
	pauseStateEnabled := true
	tdMode := "default"
	eosMode := Agora.StartAgentsRequestPropertiesTurnDetectionConfigEndOfSpeechModeSemantic

	props, err := NewAgent(
		WithGreetingConfigs(&LlmGreetingConfigs{
			Mode:          &mode,
			DelayMs:       Agora.Int(250),
			Interruptable: &interruptable,
		}),
		WithTurnDetectionConfig(&TurnDetectionConfig{
			Mode: &tdMode,
			Config: &TurnDetectionNestedConfig{
				EndOfSpeech: &EndOfSpeechConfig{
					Mode: &eosMode,
					SemanticConfig: &EndOfSpeechSemanticConfig{
						PauseStateEnabled: &pauseStateEnabled,
					},
				},
			},
		}),
	).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		APIKey: "openai-key",
		Model:  "gpt-4o-mini",
	})).WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
		APIKey: "openai-key",
		Voice:  "alloy",
	})).ToPropertiesMap(ToPropertiesOptions{
		Channel:    "room-1",
		Token:      "rtc-token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	})
	require.NoError(t, err)

	llm := props["llm"].(map[string]interface{})
	greetingConfigs := llm["greeting_configs"].(map[string]interface{})
	assert.Equal(t, "single_every", greetingConfigs["mode"])
	assert.Equal(t, float64(250), greetingConfigs["delay_ms"])
	assert.Equal(t, true, greetingConfigs["interruptable"])

	turnDetection := props["turn_detection"].(map[string]interface{})
	config := turnDetection["config"].(map[string]interface{})
	endOfSpeech := config["end_of_speech"].(map[string]interface{})
	semanticConfig := endOfSpeech["semantic_config"].(map[string]interface{})
	assert.Equal(t, true, semanticConfig["pause_state_enabled"])
}
