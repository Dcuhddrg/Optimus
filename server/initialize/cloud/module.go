package cloud

import (
	cloudApi "optimus/server/api/v1/cloud"
	"optimus/server/config"
	"optimus/server/service/cloud/infra/persistence"
	"optimus/server/service/cloud/infra/provider/aliyun"
	cloudUsecase "optimus/server/service/cloud/usecase"

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
		padded := make([]byte, 32)
		copy(padded, encKey)
		encKey = padded
	}

	registry := cloudUsecase.NewSyncerRegistry()
	registry.Register("aliyun", aliyun.NewECSSyncer())

	accountStore := persistence.NewGormAccountStore(db)
	assetStore := persistence.NewGormAssetStore(db)

	accountSvc, err := cloudUsecase.NewCloudAccountService(accountStore, encKey)
	if err != nil {
		// Key was padded to 32 bytes above, so this should never happen
		return &CloudModule{Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{})}
	}
	assetSvc := cloudUsecase.NewAssetRegistryService(accountSvc, assetStore, registry)

	return &CloudModule{
		Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{
			Account: accountSvc,
			Asset:   assetSvc,
		}),
	}
}
