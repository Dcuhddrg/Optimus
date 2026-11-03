package cloud

import (
	"testing"

	"github.com/stretchr/testify/require"
	"optimus/server/service"
	cloudService "optimus/server/service/cloud"
)

func TestRefreshServiceAliasesPublishesStartupDependencies(t *testing.T) {
	group := &service.ServiceGroupApp.CloudServiceGroup
	previous := *group
	previousAccount, previousAsset := accountService, assetService
	t.Cleanup(func() {
		*group = previous
		accountService, assetService = previousAccount, previousAsset
	})
	*group = cloudService.ServiceGroup{}
	RefreshServiceAliases()
	require.False(t, accountService.Ready())
	require.False(t, assetService.Ready())
	account, err := cloudService.NewCloudAccountService(make([]byte, 32))
	require.NoError(t, err)
	assets := cloudService.NewAssetRegistryService(account, cloudService.NewSyncerRegistry())
	group.CloudAccountService = *account
	group.AssetRegistryService = *assets
	require.False(t, accountService.Ready(), "the value alias changes only when startup publishes it")
	RefreshServiceAliases()
	require.True(t, accountService.Ready())
	require.True(t, assetService.Ready())
	*group = cloudService.ServiceGroup{}
	RefreshServiceAliases()
	require.False(t, accountService.Ready())
	require.False(t, assetService.Ready())
}
