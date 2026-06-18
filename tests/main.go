package main

import (
	"context"
	"fmt"
	"os"
	"time"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

const (
	agentPrompt = "You are a concise, technically credible voice assistant. Keep replies short unless the user asks for detail."
	greeting    = "Hi there! I am your Agora voice assistant. How can I help?"
)

func stringPtr(v string) *string    { return &v }
func intPtr(v int) *int             { return &v }
func float64Ptr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool          { return &v }

func requireEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", name)
	}
	return value, nil
}

func startConversation(ctx context.Context) (string, error) {
	appID, err := requireEnv("AGORA_APP_ID")
	if err != nil {
		return "", err
	}
	appCertificate, err := requireEnv("AGORA_APP_CERTIFICATE")
	if err != nil {
		return "", err
	}
	expiresIn, err := agentkit.ExpiresInHours(1)
	if err != nil {
		return "", err
	}

	client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          appID,
		AppCertificate: appCertificate,
	})

	agent := agentkit.NewAgent(client,
		agentkit.WithAudioScenario(agentkit.ParametersAudioScenarioChorus),
		agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
			Language: Agora.AsrLanguageEnUs.Ptr(),
			Config: &agentkit.TurnDetectionNestedConfig{
				SpeechThreshold: float64Ptr(0.5),
				StartOfSpeech: &agentkit.StartOfSpeechConfig{
					Mode: agentkit.StartOfSpeechMode("vad"),
					VadConfig: &agentkit.StartOfSpeechVadConfig{
						InterruptDurationMs: intPtr(160),
						PrefixPaddingMs:     intPtr(300),
					},
				},
				EndOfSpeech: &agentkit.EndOfSpeechConfig{
					Mode: agentkit.EndOfSpeechMode("vad").Ptr(),
					VadConfig: &agentkit.EndOfSpeechVadConfig{
						SilenceDurationMs: intPtr(480),
					},
				},
			},
		}),
		agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
			EnableRtm:   boolPtr(true),
			EnableTools: boolPtr(true),
		}),
		agentkit.WithParameters(&agentkit.SessionParams{
			DataChannel:        agentkit.DataChannelRtm.Ptr(),
			EnableErrorMessage: boolPtr(true),
		}),
	).WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
		Model:    "nova-3",
		Language: "en",
	})).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
		Model:           "gpt-4o-mini",
		SystemMessages:  []map[string]interface{}{{"role": "system", "content": agentPrompt}},
		GreetingMessage: greeting,
		FailureMessage:  "Please wait a moment.",
		MaxHistory:      intPtr(50),
		Params: map[string]interface{}{
			"max_tokens":  1024,
			"temperature": 0.7,
			"top_p":       0.95,
		},
	})).WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
		Model:   "speech_2_6_turbo",
		VoiceID: "English_captivating_female1",
	}))

	session := agent.CreateSession(agentkit.CreateSessionOptions{
		Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
		Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
		AgentUID:    "123456",
		RemoteUIDs:  []string{"*"},
		IdleTimeout: intPtr(30),
		ExpiresIn:   expiresIn,
		Debug:       true,
	})

	return session.Start(ctx)
}

func main() {
	agentID, err := startConversation(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(agentID)
}
