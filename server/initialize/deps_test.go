package initialize

import (
	"path/filepath"
	"testing"

	"optimus/server/global"

	"github.com/stretchr/testify/require"
)

func TestRedisInitSkipsWhenDisabled(t *testing.T) {
	originalConfig := global.OPTIMUS_CONFIG
	originalRedis := global.OPTIMUS_REDIS
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
		global.OPTIMUS_REDIS = originalRedis
	})

	global.OPTIMUS_CONFIG.System.UseRedis = false

	require.NotPanics(t, func() {
		Redis()
	})
}

func TestCasbinInitSkipsWhenDisabled(t *testing.T) {
	originalConfig := global.OPTIMUS_CONFIG
	originalEnforcer := global.OPTIMUS_ENFORCER
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
		global.OPTIMUS_ENFORCER = originalEnforcer
	})

	global.OPTIMUS_CONFIG.System.UseCasbin = false

	require.NotPanics(t, func() {
		Casbin()
	})
}

func TestCasbinInitLoadsModelFileWhenEnabled(t *testing.T) {
	originalConfig := global.OPTIMUS_CONFIG
	originalEnforcer := global.OPTIMUS_ENFORCER
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
		global.OPTIMUS_ENFORCER = originalEnforcer
	})

	modelPath, err := filepath.Abs(filepath.Join("..", "config", "casbin_model.conf"))
	require.NoError(t, err)

	global.OPTIMUS_CONFIG.System.UseCasbin = true
	global.OPTIMUS_CONFIG.Casbin.Enable = true
	global.OPTIMUS_CONFIG.Casbin.ModelPath = modelPath

	enforcer := Casbin()

	require.NotNil(t, enforcer)
	require.NotNil(t, global.OPTIMUS_ENFORCER)
}
