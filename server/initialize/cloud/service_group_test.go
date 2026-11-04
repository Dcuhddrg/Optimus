package cloud

import (
	"strings"
	"testing"

	"optimus/server/config"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewServiceGroupReadiness(t *testing.T) {
	unavailable := NewServiceGroup(nil, config.Config{})
	require.False(t, unavailable.CloudAccountService.Ready())
	require.False(t, unavailable.AssetRegistryService.Ready())

	for _, key := range []string{"", "short", strings.Repeat("x", 33)} {
		invalid := NewServiceGroup(&gorm.DB{}, config.Config{Cloud: config.Cloud{CredentialEncryptionKey: key}})
		require.False(t, invalid.CloudAccountService.Ready())
		require.False(t, invalid.AssetRegistryService.Ready())
	}

	available := NewServiceGroup(&gorm.DB{}, config.Config{Cloud: config.Cloud{CredentialEncryptionKey: strings.Repeat("x", 32)}})
	require.True(t, available.CloudAccountService.Ready())
	require.True(t, available.AssetRegistryService.Ready())
}
