package agentkit

import (
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseAgentForSTTLanguage() *Agent {
	return NewAgent().
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{APIKey: "llm-key", Model: "gpt-4o-mini", BaseURL: "https://api.openai.com/v1/chat/completions"})).
		WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
			Key:     "tts-key",
			VoiceID: "voice",
			ModelID: "eleven_flash_v2_5",
			BaseURL: "wss://api.elevenlabs.io/v1",
		}))
}

func propertiesForSTTLanguage(t *testing.T, agent *Agent) map[string]interface{} {
	t.Helper()

	props, err := agent.ToPropertiesMap(ToPropertiesOptions{
		Channel:    "channel",
		Token:      "token",
		AgentUID:   "1001",
		RemoteUIDs: []string{"1002"},
	})
	require.NoError(t, err)
	return props
}

func asrFromProperties(t *testing.T, props map[string]interface{}) map[string]interface{} {
	t.Helper()

	asr, ok := props["asr"].(map[string]interface{})
	require.True(t, ok)
	return asr
}

func TestSTTLanguageSerializesBCP47ToTurnDetectionAndProviderParams(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
			APIKey:   "stt-key",
			Language: "en-US",
		})))

	asr := asrFromProperties(t, props)
	assert.Equal(t, "speechmatics", asr["vendor"])
	assert.NotContains(t, asr, "language")
	assert.Equal(t, "en-US", props["turn_detection"].(map[string]interface{})["language"])
	assert.Equal(t, "en-US", asr["params"].(map[string]interface{})["language"])
}

func TestSTTProviderLanguageDefaultsTurnDetectionLanguageWhenUnsupportedByAres(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
			APIKey:   "stt-key",
			Language: "en",
		})))

	asr := asrFromProperties(t, props)
	assert.NotContains(t, asr, "language")
	assert.Equal(t, "en-US", props["turn_detection"].(map[string]interface{})["language"])
	assert.Equal(t, "en", asr["params"].(map[string]interface{})["language"])
}

func TestTurnDetectionLanguageCanDifferFromProviderLanguage(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage().
		WithTurnDetection(&TurnDetectionConfig{
			Language: Agora.AsrLanguageFrFr.Ptr(),
		}).
		WithStt(vendors.NewSpeechmaticsSTT(vendors.SpeechmaticsSTTOptions{
			APIKey:   "stt-key",
			Language: "en",
		})))

	asr := asrFromProperties(t, props)
	assert.NotContains(t, asr, "language")
	assert.Equal(t, "fr-FR", props["turn_detection"].(map[string]interface{})["language"])
	assert.Equal(t, "en", asr["params"].(map[string]interface{})["language"])
}

func TestInvalidTurnDetectionLanguagePanics(t *testing.T) {
	assert.PanicsWithValue(t, "invalid interaction language: en", func() {
		baseAgentForSTTLanguage().WithTurnDetection(&TurnDetectionConfig{
			Language: Agora.AsrLanguage("en").Ptr(),
		}).ToPropertiesMap(ToPropertiesOptions{
			Channel:    "channel",
			Token:      "token",
			AgentUID:   "1001",
			RemoteUIDs: []string{"1002"},
		})
	})
}

func TestSTTDefaultTurnDetectionLanguageIsSentWithoutSTT(t *testing.T) {
	props := propertiesForSTTLanguage(t, baseAgentForSTTLanguage())

	assert.Equal(t, map[string]interface{}{"vendor": "ares"}, props["asr"])
	assert.Equal(t, map[string]interface{}{"language": "en-US"}, props["turn_detection"])
}

func TestSTTVendorParamsMatchDocumentedShapes(t *testing.T) {
	assert.Equal(t, map[string]interface{}{
		"model":    "nova-3",
		"language": "en-US",
	}, vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{Model: "nova-3", Language: "en-US"}).ToConfig()["params"])

	assert.PanicsWithValue(t, "DeepgramSTT requires APIKey unless using a supported Agora-managed model", func() {
		vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{Model: "enhanced"})
	})

	assert.Equal(t, map[string]interface{}{
		"key":      "dg-key",
		"language": "en",
	}, vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{APIKey: "dg-key", Language: "en"}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"api_key": "openai-key",
		"input_audio_transcription": map[string]interface{}{
			"model":    "gpt-4o-mini-transcribe",
			"language": "en",
		},
	}, vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
		APIKey:   "openai-key",
		Model:    "gpt-4o-mini-transcribe",
		Language: "en",
	}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"api_key": "openai-key",
		"input_audio_transcription": map[string]interface{}{
			"model": "whisper-1",
		},
	}, vendors.NewOpenAISTT(vendors.OpenAISTTOptions{
		APIKey: "openai-key",
	}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"project_id":             "project",
		"location":               "global",
		"adc_credentials_string": "{}",
		"language":               "en-US",
		"model":                  "long",
	}, vendors.NewGoogleSTT(vendors.GoogleSTTOptions{
		ProjectID:            "project",
		Location:             "global",
		ADCCredentialsString: "{}",
		Language:             "en-US",
		Model:                "long",
	}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"access_key_id":     "access",
		"secret_access_key": "secret",
		"region":            "us-east-1",
		"language_code":     "en-US",
	}, vendors.NewAmazonSTT(vendors.AmazonSTTOptions{
		AccessKey: "access",
		SecretKey: "secret",
		Region:    "us-east-1",
		Language:  "en-US",
	}).ToConfig()["params"])

	assert.Equal(t, map[string]interface{}{
		"api_key":  "assembly-key",
		"language": "en-US",
		"uri":      "wss://example.test/ws",
	}, vendors.NewAssemblyAISTT(vendors.AssemblyAISTTOptions{
		APIKey:   "assembly-key",
		Language: "en-US",
		URI:      "wss://example.test/ws",
	}).ToConfig()["params"])
}
