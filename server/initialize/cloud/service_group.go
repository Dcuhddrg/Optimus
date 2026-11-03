package cloud

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"optimus/server/config"
	"optimus/server/global"
	cloudService "optimus/server/service/cloud"
	"optimus/server/service/cloud/provider/aliyun"
)

// NewServiceGroup assembles all dependencies for the shared cloud module group.
func NewServiceGroup(db *gorm.DB, cfg config.Config) cloudService.ServiceGroup {
	if db == nil {
		return cloudService.ServiceGroup{}
	}

	encKey := []byte(cfg.Cloud.CredentialEncryptionKey)
	if len(encKey) != 32 {
		if logger, ok := global.OPTIMUS_LOG.(*zap.Logger); ok && logger != nil {
			logger.Warn("cloud credential encryption key is not 32 bytes; using zero-padded fallback — set cloud.credential-encryption-key in config before storing credentials",
				zap.Int("configured_len", len(encKey)))
		}
		padded := make([]byte, 32)
		copy(padded, encKey)
		encKey = padded
	}

	registry := cloudService.NewSyncerRegistry()
	registry.Register("aliyun", aliyun.NewECSSyncer())

	accountSvc, err := cloudService.NewCloudAccountService(encKey)
	if err != nil {
		// Key was padded to 32 bytes above, so this should never happen
		return cloudService.ServiceGroup{}
	}
	assetSvc := cloudService.NewAssetRegistryService(accountSvc, registry)

	return cloudService.ServiceGroup{
		CloudAccountService:  *accountSvc,
		AssetRegistryService: *assetSvc,
	}
}
