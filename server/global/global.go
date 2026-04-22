package global

import (
	"optimus/server/config"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	OPTIMUS_CONFIG   config.Config
	OPTIMUS_VP       *viper.Viper
	OPTIMUS_LOG      any
	OPTIMUS_DB       *gorm.DB
	OPTIMUS_REDIS    *redis.Client
	OPTIMUS_ENFORCER *casbin.Enforcer
)
