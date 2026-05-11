package initialize

import (
	"testing"

	"optimus/server/config"
	"optimus/server/global"

	"gorm.io/gorm"

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
	require.Equal(t, cfg, module.cfg)
	require.NotNil(t, module.Handlers)
	require.NotNil(t, module.Handlers.Auth)
	require.NotNil(t, module.Handlers.Base)
	require.NotNil(t, module.Handlers.Health)
	require.NotNil(t, module.Capabilities)
	require.Nil(t, module.Capabilities.RBAC)
}

func TestNewSystemModuleBuildsRBACCapabilityWhenDatabaseExists(t *testing.T) {
	cfg := config.Config{}

	module := NewSystemModule(&gorm.DB{}, cfg)

	require.NotNil(t, module)
	require.NotNil(t, module.Capabilities)
	require.NotNil(t, module.Capabilities.RBAC)
}
