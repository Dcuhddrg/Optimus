package cloud

import (
	"testing"

	"optimus/server/global"
)

// useTestCredentialKey sets a valid 32-byte key in global config for one test.
func useTestCredentialKey(t testing.TB) {
	t.Helper()
	saved := global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey
	t.Cleanup(func() { global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey = saved })
	global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey = string(make([]byte, 32))
}

func newCloudAccountServiceForTest(t testing.TB) (*CloudAccountService, error) {
	t.Helper()
	useTestCredentialKey(t)
	return &CloudAccountService{}, nil
}

// newAssetRegistryServiceForTest swaps the process-wide syncer registry for one test.
func newAssetRegistryServiceForTest(t testing.TB, registry *SyncerRegistry) *AssetRegistryService {
	t.Helper()
	saved := Syncers
	t.Cleanup(func() { Syncers = saved })
	Syncers = registry
	return &AssetRegistryService{}
}
