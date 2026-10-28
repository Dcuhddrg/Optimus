package cloud

import (
	cloudApi "optimus/server/api/v1/cloud"
	"optimus/server/config"
	"optimus/server/global"
	cloudService "optimus/server/service/cloud"
	"optimus/server/service/cloud/provider/aliyun"
	cloudStore "optimus/server/service/cloud/store"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CloudModule holds all dependencies for the cloud module.
type CloudModule struct {
	Handlers *cloudApi.CloudHandlers
}

func NewCloudModule(db *gorm.DB, cfg config.Config) *CloudModule {
	if db == nil {
		return &CloudModule{Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{})}
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

	accountStore := cloudStore.NewGormAccountStore(db)
	assetStore := cloudStore.NewGormAssetStore(db)

	accountSvc, err := cloudService.NewCloudAccountService(accountStore, encKey)
	if err != nil {
		// Key was padded to 32 bytes above, so this should never happen
		return &CloudModule{Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{})}
	}
	assetSvc := cloudService.NewAssetRegistryService(accountSvc, assetStore, registry)

	return &CloudModule{
		Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{
			Account: accountSvc,
			Asset:   assetSvc,
		}),
	}
}
