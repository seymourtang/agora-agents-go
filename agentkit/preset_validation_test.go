package agentkit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/client"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startPresetValidationSession(t *testing.T, agent *Agent, opts CreateSessionOptions) (map[string]interface{}, error) {
	t.Helper()

	httpClient := &captureStartHTTPClient{}
	rawClient := client.NewClient(
		option.WithBaseURL("https://api.example.test"),
		option.WithHTTPClient(httpClient),
	)
	agoraClient := &AgoraClient{
		Agents: rawClient.Agents,
		AppID:  "appid",
	}

	session := agent.CreateSession(agoraClient, opts)
	_, err := session.Start(context.Background())
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(httpClient.lastBody, &payload))
	return payload, nil
}

func TestExplicitASRPresetStillRequiresTTSAndLLM(t *testing.T) {
	opts := basePipelineSessionOptions()
	opts.Preset = []string{AgentPresets.Asr.DeepgramNova3}

	_, err := startPresetValidationSession(t, NewAgent(WithName("support")), opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TTS configuration is required")
}

func TestExplicitLLMPresetStillRequiresTTS(t *testing.T) {
	opts := basePipelineSessionOptions()
	opts.Preset = []string{AgentPresets.Llm.OpenAIGpt4oMini}

	_, err := startPresetValidationSession(t, NewAgent(WithName("support")), opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TTS configuration is required")
}

func TestExplicitTTSPresetStillRequiresLLM(t *testing.T) {
	opts := basePipelineSessionOptions()
	opts.Preset = []string{AgentPresets.Tts.OpenAITts1}

	_, err := startPresetValidationSession(t, NewAgent(WithName("support")), opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "LLM configuration is required")
}

func TestSessionStartInfersASRLLMAndTTSPresets(t *testing.T) {
	agent := NewAgent(WithName("support")).
		WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{Model: "nova-3", Language: "en-US"})).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
		WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{Voice: "alloy"}))

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)

	assert.Equal(t, "deepgram_nova_3,openai_gpt_4o_mini,openai_tts_1", payload["preset"])
	properties := payload["properties"].(map[string]interface{})
	assert.Equal(t, map[string]interface{}{"language": "en-US"}, properties["asr"].(map[string]interface{})["params"])
	assert.Equal(t, map[string]interface{}{"voice": "alloy"}, properties["tts"].(map[string]interface{})["params"])
}

func TestSessionStartInfersHyphenatedMiniMaxManagedPresetModel(t *testing.T) {
	agent := NewAgent(WithName("support")).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{Model: "gpt-4o-mini"})).
		WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Model:   "speech-2.6-turbo",
			VoiceID: "English_captivating_female1",
		}))

	payload, err := startPresetValidationSession(t, agent, basePipelineSessionOptions())
	require.NoError(t, err)

	assert.Equal(t, "openai_gpt_4o_mini,minimax_speech_2_6_turbo", payload["preset"])
	properties := payload["properties"].(map[string]interface{})
	assert.Equal(
		t,
		map[string]interface{}{"voice_id": "English_captivating_female1"},
		properties["tts"].(map[string]interface{})["params"].(map[string]interface{})["voice_setting"],
	)
}

func TestMiniMaxSpeech02TurboRequiresBYOK(t *testing.T) {
	assert.PanicsWithValue(t, "MiniMaxTTS requires Key unless using a supported Agora-managed model", func() {
		vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Model:   "speech-02-turbo",
			VoiceID: "English_captivating_female1",
		})
	})
}
