package agentkit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAgentRequiresClient(t *testing.T) {
	require.PanicsWithValue(t, "NewAgent requires AgoraClient", func() {
		NewAgent(nil)
	})
}

func TestValidateAvatarConfigAllowsSensetimeWithoutSceneList(t *testing.T) {
	err := ValidateAvatarConfig("sensetime", map[string]interface{}{
		"agora_uid": "2001",
		"appId":     "sensetime-app",
		"app_key":   "sensetime-key",
	})
	require.NoError(t, err)
}
