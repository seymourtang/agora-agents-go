package agentkit

import (
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAudioScenarioWhenUnset(t *testing.T) {
	props, err := NewAgent(testAgoraClient()).
		WithLlm(stubLLM).
		WithTts(stubTTS).
		ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	params := props["parameters"].(map[string]interface{})
	assert.Equal(t, "default", params["audio_scenario"])
}

func TestAudioScenarioRespectsExplicitValue(t *testing.T) {
	props, err := NewAgent(testAgoraClient(),
		WithAudioScenario(ParametersAudioScenarioChorus),
	).WithLlm(stubLLM).
		WithTts(stubTTS).
		ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	params := props["parameters"].(map[string]interface{})
	assert.Equal(t, "chorus", params["audio_scenario"])
}

func TestDefaultAudioScenarioForMLLM(t *testing.T) {
	props, err := NewAgent(testAgoraClient()).
		WithMllm(vendors.NewOpenAIRealtime(vendors.OpenAIRealtimeOptions{
			APIKey: "rt-key",
			Model:  "gpt-4o-realtime-preview",
		})).
		ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	params := props["parameters"].(map[string]interface{})
	assert.Equal(t, "default", params["audio_scenario"])
}
