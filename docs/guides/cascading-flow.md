---
sidebar_position: 1
title: Cascading Flow
description: Build agents using the ASR -> LLM -> TTS cascading flow with different vendor combinations.
---

# Cascading Flow (ASR -> LLM -> TTS)

The cascading flow chains three vendor services: speech-to-text (STT/ASR) converts audio to text, an LLM generates a response, and text-to-speech (TTS) converts it back to audio.

## Example 1: OpenAI + ElevenLabs + Deepgram

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    Agora "github.com/AgoraIO/agora-agents-go/v2"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "<app_id>",
        AppCertificate: "<app_cert>",
    })

    agent := agentkit.NewAgent(client).WithLlm(
        vendors.NewOpenAI(vendors.OpenAIOptions{
            APIKey:  "<openai_key>",
            BaseURL: "https://api.openai.com/v1/chat/completions",
            Model:   "gpt-4o-mini",
            SystemMessages: []map[string]interface{}{
                {"role": "system", "content": "You are a helpful voice assistant. Keep responses under 3 sentences."},
            },
            GreetingMessage: "Hi there! What can I help you with?",
            MaxHistory:      Agora.Int(20),
        }),
    ).WithTts(
        vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
            Key:     "<elevenlabs_key>",
            ModelID:    "eleven_turbo_v2_5",
            VoiceID:    "<voice_id>",
            BaseURL:    "wss://api.elevenlabs.io/v1",
        }),
    ).WithStt(
        vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
            APIKey:   "<deepgram_key>",
            Model:    "nova-2",
            Language: "en-US",
        }),
    )

    session := agent.CreateSession(agentkit.CreateSessionOptions{
        Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
        Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    ctx := context.Background()

    agentID, err := session.Start(ctx)
    if err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    fmt.Println("Running:", agentID)

    // ... interact with the agent ...

    err = session.Stop(ctx)
    if err != nil {
        log.Fatalf("Failed to stop: %v", err)
    }
}
```

## Example 2: Anthropic + Microsoft TTS + Microsoft STT (App Credentials)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    Agora "github.com/AgoraIO/agora-agents-go/v2"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "<app_id>",
        AppCertificate: "<app_cert>",
    })

    agent := agentkit.NewAgent(client).WithLlm(
        vendors.NewAnthropic(vendors.AnthropicOptions{
            APIKey:    "<anthropic_key>",
            URL:       "https://api.anthropic.com/v1/messages",
            Headers:   map[string]string{"anthropic-version": "2023-06-01"},
            Model:     "claude-3-5-sonnet-20241022",
            MaxTokens: Agora.Int(1024),
            SystemMessages: []map[string]interface{}{
                {"role": "system", "content": "You are a customer service agent for Acme Corp."},
            },
            GreetingMessage: "Welcome to Acme Corp! How may I assist you?",
            FailureMessage:  "I apologize for the inconvenience. Please try again shortly.",
        }),
    ).WithTts(
        vendors.NewMicrosoftTTS(vendors.MicrosoftTTSOptions{
            Key:       "<microsoft_key>",
            Region:    "eastus",
            VoiceName: "en-US-JennyNeural",
        }),
    ).WithStt(
        vendors.NewMicrosoftSTT(vendors.MicrosoftSTTOptions{
            Key:    "<microsoft_key>",
            Region: "eastus",
        }),
    )

    session := agent.CreateSession(agentkit.CreateSessionOptions{
        Name:        fmt.Sprintf("conversation-%d", time.Now().UnixMilli()),
        Channel:     fmt.Sprintf("demo-channel-%d", time.Now().UnixMilli()),
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    ctx := context.Background()

    agentID, err := session.Start(ctx)
    if err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    fmt.Println("Running:", agentID)

    err = session.Stop(ctx)
    if err != nil {
        log.Fatalf("Failed to stop: %v", err)
    }
}
```

## Required Vendors

In cascading mode, **LLM and TTS are required**. STT is optional — if omitted, the platform uses a default ASR provider. `ToProperties` returns an error if LLM or TTS is missing:

```go
agent := agentkit.NewAgent(client)
// No TTS or LLM configured

_, err := agent.ToProperties(agentkit.ToPropertiesOptions{...})
// err: "TTS configuration is required; use WithTts() to set it"
```

## Turn Detection

Add server-side voice activity detection to control when the agent starts processing:

```go
import Agora "github.com/AgoraIO/agora-agents-go/v2"

agent := agentkit.NewAgent(client,
    agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
        Type:              agentkit.TurnDetectionTypeServerVad.Ptr(), // deprecated; use Config.EndOfSpeech instead
        Threshold:         Agora.Float64(0.5),
        SilenceDurationMs: Agora.Int(500),
    }),
    // ... other options
)
```
