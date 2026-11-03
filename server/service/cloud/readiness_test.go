package cloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudServiceReadinessReflectsConstruction(t *testing.T) {
	require.False(t, (&CloudAccountService{}).Ready())
	require.False(t, (&AssetRegistryService{}).Ready())
	account, err := NewCloudAccountService(make([]byte, 32))
	require.NoError(t, err)
	require.True(t, account.Ready())
	assets := NewAssetRegistryService(account, NewSyncerRegistry())
	require.True(t, assets.Ready())
	// Construction previously made the handler available even with nil dependencies.
	require.True(t, NewAssetRegistryService(nil, nil).Ready())
}
