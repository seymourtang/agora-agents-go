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

func startPipelineIDSession(t *testing.T, agent *Agent, opts CreateSessionOptions) map[string]interface{} {
	t.Helper()

	httpClient := &captureStartHTTPClient{}
	rawClient := client.NewClient(
		option.WithBaseURL("https://api.example.test"),
		option.WithHTTPClient(httpClient),
	)
	agoraClient := &AgoraClient{
		Agents: rawClient.Agents,
		appID:  "appid",
	}

	agent.base.Client = agoraClient
	session := agent.CreateSession(opts)
	agentID, err := session.Start(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "agent_123", agentID)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(httpClient.lastBody, &payload))
	return payload
}

func basePipelineSessionOptions() CreateSessionOptions {
	return CreateSessionOptions{
		Channel:    "channel",
		Token:      "token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	}
}

func TestAgentPipelineIDSendsTopLevelPipelineID(t *testing.T) {
	payload := startPipelineIDSession(t, NewAgent(nil, WithName("support"), WithPipelineID("studio-pipeline-id")), basePipelineSessionOptions())

	assert.Equal(t, "support", payload["name"])
	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.Equal(t, "channel", properties["channel"])
	assert.Equal(t, "token", properties["token"])
	assert.Equal(t, "1", properties["agent_rtc_uid"])
	assert.NotContains(t, properties, "pipeline_id")
}

func TestSessionPipelineIDOverridesAgentPipelineID(t *testing.T) {
	opts := basePipelineSessionOptions()
	opts.PipelineID = "session-pipeline"

	payload := startPipelineIDSession(t, NewAgent(nil, WithName("support"), WithPipelineID("agent-pipeline")), opts)

	assert.Equal(t, "session-pipeline", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "pipeline_id")
}

func TestAgentPipelineIDSkipsMissingVendorValidation(t *testing.T) {
	payload := startPipelineIDSession(t, NewAgent(nil, WithName("support"), WithPipelineID("studio-pipeline-id")), basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "asr")
	assert.NotContains(t, properties, "llm")
	assert.NotContains(t, properties, "tts")
}

func TestPipelineIDAllowsSingleLLMOverrideWithoutTTSOrASR(t *testing.T) {
	agent := NewAgent(nil, WithName("support"), WithPipelineID("studio-pipeline-id")).WithLlm(
		vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey:  "openai-key",
			BaseURL: "https://api.openai.com/v1/chat/completions",
			Model:   "gpt-4o",
		}),
	)

	payload := startPipelineIDSession(t, agent, basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "asr")
	assert.NotContains(t, properties, "tts")
	llm := properties["llm"].(map[string]interface{})
	assert.Equal(t, "openai-key", llm["api_key"])
	assert.Equal(t, "gpt-4o", llm["params"].(map[string]interface{})["model"])
}

func TestPipelineIDAllowsMultipleOverridesWithoutASR(t *testing.T) {
	agent := NewAgent(nil, WithName("support"), WithPipelineID("studio-pipeline-id")).
		WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
			APIKey:  "openai-key",
			BaseURL: "https://api.openai.com/v1/chat/completions",
			Model:   "gpt-4o",
		})).
		WithTts(vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
			APIKey:  "tts-key",
			BaseURL: "https://api.openai.com/v1/audio/speech",
			Model:   "tts-1-hd",
			Voice:   "alloy",
		}))

	payload := startPipelineIDSession(t, agent, basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "asr")
	assert.Equal(t, "openai-key", properties["llm"].(map[string]interface{})["api_key"])
	tts := properties["tts"].(map[string]interface{})
	assert.Equal(t, "openai", tts["vendor"])
	assert.Equal(t, "tts-key", tts["params"].(map[string]interface{})["api_key"])
}

func TestAgentPipelineIDSurvivesBuilderClone(t *testing.T) {
	agent := NewAgent(nil, WithName("support"), WithPipelineID("studio-pipeline-id")).WithTools(true)

	assert.Equal(t, "studio-pipeline-id", agent.PipelineID())
	payload := startPipelineIDSession(t, agent, basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	advancedFeatures := properties["advanced_features"].(map[string]interface{})
	assert.Equal(t, true, advancedFeatures["enable_tools"])
}
