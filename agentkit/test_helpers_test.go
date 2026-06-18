package agentkit

import "github.com/AgoraIO/agora-agents-go/v2/option"

func testAgoraClient() *AgoraClient {
	return NewAgoraClient(AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          "test-app-id",
		AppCertificate: "test-app-certificate",
	})
}
