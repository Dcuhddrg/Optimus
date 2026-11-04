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
	accountSvc, err := cloudService.NewCloudAccountService(encKey)
	if err != nil {
		if logger, ok := global.OPTIMUS_LOG.(*zap.Logger); ok && logger != nil {
			logger.Error("cloud credential encryption key is invalid; cloud services are unavailable",
				zap.Int("configured_len", len(encKey)), zap.Error(err))
		}
		return cloudService.ServiceGroup{}
	}

	registry := cloudService.NewSyncerRegistry()
	registry.Register("aliyun", aliyun.NewECSSyncer())
	assetSvc := cloudService.NewAssetRegistryService(accountSvc, registry)

	return cloudService.ServiceGroup{
		CloudAccountService:  *accountSvc,
		AssetRegistryService: *assetSvc,
	}
}
