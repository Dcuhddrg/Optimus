package cloud

import (
	"testing"

	"optimus/server/config"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewServiceGroupReadiness(t *testing.T) {
	unavailable := NewServiceGroup(nil, config.Config{})
	require.False(t, unavailable.CloudAccountService.Ready())
	require.False(t, unavailable.AssetRegistryService.Ready())

	available := NewServiceGroup(&gorm.DB{}, config.Config{})
	require.True(t, available.CloudAccountService.Ready())
	require.True(t, available.AssetRegistryService.Ready())
}
