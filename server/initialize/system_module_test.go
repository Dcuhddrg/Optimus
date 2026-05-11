package initialize

import (
	"testing"

	"optimus/server/global"

	"github.com/stretchr/testify/require"
)

func TestNewSystemModuleBuildsPublicHandlersWithoutDatabase(t *testing.T) {
	originalConfig := global.OPTIMUS_CONFIG
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
	})

	cfg := originalConfig
	module := NewSystemModule(nil, cfg)

	require.NotNil(t, module)
	require.NotNil(t, module.Auth)
	require.NotNil(t, module.Base)
	require.NotNil(t, module.Health)
}
