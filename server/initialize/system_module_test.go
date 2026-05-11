package initialize

import (
	"testing"

	"optimus/server/global"

	"github.com/stretchr/testify/require"
)

func TestNewSystemModuleBuildsPublicHandlersWithoutDatabase(t *testing.T) {
	module := NewSystemModule(nil, global.OPTIMUS_CONFIG)

	require.NotNil(t, module)
	require.NotNil(t, module.Auth)
	require.NotNil(t, module.Base)
	require.NotNil(t, module.Health)
}
