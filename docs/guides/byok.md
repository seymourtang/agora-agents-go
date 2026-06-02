---
sidebar_position: 4
title: BYOK
description: Bring your own vendor credentials and use custom vendor configuration with the Go SDK.
---

# BYOK

Use BYOK when you want to provide vendor credentials yourself instead of using Agora-managed models.

Typical reasons:

- you need a vendor model that is not part of the Agora-managed catalog
- you want to point to a custom endpoint
- you want direct control over vendor-specific parameters
- your organization manages vendor billing separately from Agora

## Full example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/AgoraIO/agora-agents-go/v2/agentkit"
    "github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
    "github.com/AgoraIO/agora-agents-go/v2/option"
)

func main() {
    ctx := context.Background()
    idleTimeout := 120

    client := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
        Area:           option.AreaUS,
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })

    // In BYOK mode, each vendor carries its own credentials.
    sampleRate := vendors.SampleRate24kHz
    agent := agentkit.NewAgent(
        agentkit.WithName("support-assistant"),
    ).WithStt(vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
        APIKey:   os.Getenv("DEEPGRAM_API_KEY"),
        Model:    "nova-3",
        Language: "en-US",
    })).WithLlm(vendors.NewOpenAI(vendors.OpenAIOptions{
        APIKey:  os.Getenv("OPENAI_API_KEY"),
        BaseURL: "https://api.openai.com/v1/chat/completions",
        Model:   "gpt-4o-mini",
        SystemMessages: []map[string]interface{}{
            {"role": "system", "content": "You are a concise support voice assistant."},
        },
        GreetingMessage: "Hello! How can I help you today?",
        MaxHistory:      intPtr(10),
    })).WithTts(vendors.NewElevenLabsTTS(vendors.ElevenLabsTTSOptions{
        Key:        os.Getenv("ELEVENLABS_API_KEY"),
        ModelID:    "eleven_flash_v2_5",
        VoiceID:    os.Getenv("ELEVENLABS_VOICE_ID"),
        BaseURL:    "wss://api.elevenlabs.io/v1",
        SampleRate: &sampleRate,
    }))

    session := agent.CreateSession(client, agentkit.CreateSessionOptions{
        Channel:     "support-room-123",
        AgentUID:    "1",
        RemoteUIDs:  []string{"100"},
        IdleTimeout: &idleTimeout,
    })

    agentSessionID, err := session.Start(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Agent started:", agentSessionID)

    if err := session.Stop(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Agora-managed models vs BYOK

- Agora-managed models: supported models without vendor keys in app code
- BYOK: most control, your keys and your vendor configuration
