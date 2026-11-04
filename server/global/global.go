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
	// OPTIMUS_STARTUP_ERR is set once before the HTTP server starts when the
	// database or data initialization failed; /health then reports unhealthy.
	OPTIMUS_STARTUP_ERR error
)
