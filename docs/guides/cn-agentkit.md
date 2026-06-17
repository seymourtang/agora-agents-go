---
sidebar_position: 5
title: CN AgentKit
description: Use the CN AgentKit facade for Chinese mainland API routing.
---

# CN AgentKit

Use the CN AgentKit facade when your integration should route to Chinese mainland API endpoints (`option.AreaCN`).

```go
package main

import (
    "context"

    agentkit "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn"
    vendors "github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors"
)

func main() {
    client := agentkit.NewAgoraClient(agentkit.ClientOptions{
        AppID:          "your-app-id",
        AppCertificate: "your-app-certificate",
    })

    agent := agentkit.NewAgent(client,
        agentkit.WithName("cn-support"),
    ).WithStt(
        vendors.NewTencentSTT(vendors.TencentSTTOptions{
            Key:    "secret-key",
            AppID:  "app-id",
            Secret: "secret",
        }),
    ).WithLlm(
        vendors.NewTencentLLM(vendors.TencentLLMOptions{
            APIKey:  "api-key",
            Model:   "hunyuan-turbos-latest",
            BaseURL: "https://api.hunyuan.cloud.tencent.com/v1/chat/completions",
        }),
    ).WithTts(
        vendors.NewTencentTTS(vendors.TencentTTSOptions{
            AppID:     "app-id",
            SecretID:  "secret-id",
            SecretKey: "secret-key",
        }),
    )

    session := agent.CreateSession(agentkit.CreateSessionOptions{
        Channel:    "cn-room",
        AgentUID:   "1001",
        RemoteUIDs: []string{"1002"},
    })

    _, _ = session.Start(context.Background())
}
```

## Routing Behavior

`agentkit/cn.NewAgoraClient(...)` always uses `option.AreaCN` internally. You do not need to pass an `Area` value.

## Package Layout

- Global/default package: `github.com/AgoraIO/agora-agents-go/v2/agentkit`
- Global/default vendors: `github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors`
- CN facade: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn`
- CN vendors: `github.com/AgoraIO/agora-agents-go/v2/agentkit/cn/vendors`

## Supported Surfaces

The CN facade currently supports:

- Cascading ASR / LLM / TTS flows
- CN-specific STT, LLM, TTS, and avatar vendors
- CN-specific STT vendors such as Fengming, Tencent, Microsoft, and Xfyun are exposed from `agentkit/cn/vendors`
- The CN avatar surface currently exposes Sensetime only

For CN TTS vendors, `AdditionalParams map[string]interface{}` and `SkipPatterns []int` are available consistently across all TTS option structs as shared extension fields.

The CN facade does not expose MLLM helpers.
