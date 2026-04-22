package global

import (
	"optimus/server/config"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	OPTIMUS_CONFIG   config.Config
	OPTIMUS_VP       *viper.Viper
	OPTIMUS_LOG      any
	OPTIMUS_DB       *gorm.DB
	OPTIMUS_REDIS    any
	OPTIMUS_ENFORCER any
)
