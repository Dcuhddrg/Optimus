package initialize

import (
	"testing"

	"optimus/server/config"
	"optimus/server/global"

	"github.com/stretchr/testify/require"
)

func TestNewSystemModuleBuildsPublicHandlersWithoutDatabase(t *testing.T) {
	originalConfig := global.OPTIMUS_CONFIG
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
	})

	cfg := config.Config{
		System: config.System{
			Env:          "test",
			RouterPrefix: "/local-prefix",
		},
		JWT: config.JWT{
			SigningKey:  "local-signing-key",
			ExpiresTime: 60,
			BufferTime:  10,
			Issuer:      "local-issuer",
		},
	}
	global.OPTIMUS_CONFIG = config.Config{
		System: config.System{
			Env:          "global",
			RouterPrefix: "/global-prefix",
		},
		JWT: config.JWT{
			SigningKey:  "global-signing-key",
			ExpiresTime: 1,
			BufferTime:  1,
			Issuer:      "global-issuer",
		},
	}
	module := NewSystemModule(nil, cfg)

	require.NotNil(t, module)
	require.NotNil(t, module.Auth)
	require.NotNil(t, module.Base)
	require.NotNil(t, module.Health)
}
