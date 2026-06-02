package agentkit

import (
	"context"
	"encoding/json"
	"testing"

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
		AppID:  "appid",
	}

	session := agent.CreateSession(agoraClient, opts)
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
	payload := startPipelineIDSession(t, NewAgent(WithName("support"), WithPipelineID("studio-pipeline-id")), basePipelineSessionOptions())

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

	payload := startPipelineIDSession(t, NewAgent(WithName("support"), WithPipelineID("agent-pipeline")), opts)

	assert.Equal(t, "session-pipeline", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	assert.NotContains(t, properties, "pipeline_id")
}

func TestAgentPipelineIDSkipsMissingVendorValidation(t *testing.T) {
	payload := startPipelineIDSession(t, NewAgent(WithName("support"), WithPipelineID("studio-pipeline-id")), basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
}

func TestAgentPipelineIDSurvivesBuilderClone(t *testing.T) {
	agent := NewAgent(WithName("support"), WithPipelineID("studio-pipeline-id")).WithTools(true)

	assert.Equal(t, "studio-pipeline-id", agent.PipelineID())
	payload := startPipelineIDSession(t, agent, basePipelineSessionOptions())

	assert.Equal(t, "studio-pipeline-id", payload["pipeline_id"])
	properties := payload["properties"].(map[string]interface{})
	advancedFeatures := properties["advanced_features"].(map[string]interface{})
	assert.Equal(t, true, advancedFeatures["enable_tools"])
}
