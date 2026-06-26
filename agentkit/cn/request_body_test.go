package cn

import (
	"testing"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAgoraClient() *AgoraClient {
	return NewAgoraClient(ClientOptions{
		AppID:          "test-app-id",
		AppCertificate: "test-app-certificate",
	})
}

func basePropertiesOpts() ToPropertiesOptions {
	return ToPropertiesOptions{
		Channel:    "channel",
		Token:      "token",
		AgentUID:   "1",
		RemoteUIDs: []string{"100"},
	}
}

func TestDefaultASRFallsBackToFengming(t *testing.T) {
	agent := NewAgent(testAgoraClient()).
		WithLlm(vendors.NewAliyun(vendors.AliyunOptions{
			APIKey:  "aliyun-key",
			Model:   "qwen-max",
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		})).
		WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
			Key:   "mm-key",
			Model: "speech-2.6-turbo",
			VoiceSetting: &vendors.MiniMaxVoiceSetting{
				VoiceID: "Chinese (Mandarin)_Cheerful_Female",
			},
		}))

	props, err := agent.ToPropertiesMap(basePropertiesOpts())
	require.NoError(t, err)

	asr := props["asr"].(map[string]interface{})
	assert.Equal(t, "fengming", asr["vendor"])
	assert.Equal(t, "en-US", asr["language"])
	assert.NotContains(t, asr, "params")
}
