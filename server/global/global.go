package global

import (
	"optimus/server/config"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"github.com/spf13/viper"
	"golang.org/x/sync/singleflight"
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

	// OPTIMUS_Concurrency_Control merges concurrent token refreshes, as GVA.
	OPTIMUS_Concurrency_Control = &singleflight.Group{}
	// BlackCache holds revoked tokens until they expire, as GVA. OtherInit
	// replaces this default with one expiring after the token lifetime.
	BlackCache = local_cache.NewCache()
)
