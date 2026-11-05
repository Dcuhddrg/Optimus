package initialize

import (
	"optimus/server/global"
	cloudService "optimus/server/service/cloud"
	"optimus/server/service/cloud/provider/aliyun"

	"go.uber.org/zap"
)

// Cloud registers asset syncers and reports an unusable credential key. Cloud
// APIs stay unavailable until the key is exactly 32 bytes.
func Cloud() {
	cloudService.Syncers.Register("aliyun", aliyun.NewECSSyncer())

	if _, err := cloudService.CredentialKey(); err != nil {
		if logger := global.OPTIMUS_LOG; logger != nil {
			logger.Error("cloud credential encryption key is invalid; cloud services are unavailable",
				zap.Int("configured_len", len(global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey)), zap.Error(err))
		}
	}
}
