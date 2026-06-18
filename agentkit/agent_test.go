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
