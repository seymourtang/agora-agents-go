---
sidebar_position: 3
title: Quick Start
description: Build and run your first Agora Conversational AI agent in Go with app credentials and the builder API.
---

# Quick Start

This guide starts with the standard AgentKit path:

- `AppID`, `AppCertificate`, and `Area` on `AgoraClient`
- the `Agent` builder with `WithStt()`, `WithLlm()`, and `WithTts()`
- automatic ConvoAI REST auth and RTC join token generation
- no vendor API keys when using supported Agora-managed global models

## Full example

```go
package main

import (
    "context"
    "fmt"
    "log"

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

    agent := agentkit.NewAgent(client,
        agentkit.WithName("support-assistant"),
    ).WithStt(
        vendors.NewDeepgramSTT(vendors.DeepgramSTTOptions{
            Model:    "nova-3",
            Language: "en",
        }),
    ).WithLlm(
        vendors.NewOpenAI(vendors.OpenAIOptions{
            Model: "gpt-5-mini",
            SystemMessages: []map[string]interface{}{
                {"role": "system", "content": "You are a concise support voice assistant."},
            },
            GreetingMessage: "Hello! How can I help you today?",
            MaxHistory:      intPtr(10),
        }),
    ).WithTts(
        vendors.NewOpenAITTS(vendors.OpenAITTSOptions{
            Model: "tts-1",
            Voice: "alloy",
        }),
    )

    session := agent.CreateSession(agentkit.CreateSessionOptions{
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

    if err := session.Say(ctx, "Thanks for calling Agora support.", nil, nil); err != nil {
        log.Fatal(err)
    }
    if err := session.Stop(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## What this does

1. `AgoraClient` runs in app-credentials mode when you pass `AppID` and `AppCertificate` only.
2. `Agent` holds reusable behavior such as instructions, greeting, and history settings.
3. Vendor constructors on the builder select the ASR, LLM, and TTS stack. Leave vendor credentials unset for supported Agora-managed global models, or provide keys when you want BYOK. CN MiniMax TTS always requires `Key`.
4. `session.Start(...)` lets the SDK generate the required auth tokens automatically.
5. `session.Start(...)` returns the unique agent session ID.

## When to use BYOK instead

Use the builder without vendor API keys when you are using supported Agora-managed global models.

Use BYOK when you need to:

- supply your own vendor API keys
- use models outside the Agora-managed catalog
- point at custom vendor endpoints
- manage vendor-specific parameters directly

See [BYOK Guide](../guides/byok.md).

## Next steps

- [Authentication](./authentication.md)
- [BYOK Guide](../guides/byok.md)
- [MLLM Flow](../guides/mllm-flow.md)
- [Agent Reference](../reference/agent.md)
