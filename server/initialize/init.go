package initialize

import (
	"strings"

	"optimus/server/global"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/songzhibin97/gkit/cache/local_cache"
)

// OtherInit validates JWT lifetimes and creates the token blacklist cache, as GVA.
func OtherInit() {
	dr, err := utils.ParseDuration(global.OPTIMUS_CONFIG.JWT.ExpiresTime)
	if err != nil {
		panic(err)
	}
	if _, err = utils.ParseDuration(global.OPTIMUS_CONFIG.JWT.BufferTime); err != nil {
		panic(err)
	}
	global.BlackCache = local_cache.NewCache(local_cache.SetDefaultExpire(dr))
	gin.SetMode(ginMode(global.OPTIMUS_CONFIG.System.Env))
	setupValidator()
}

func ginMode(env string) string {
	switch strings.ToLower(env) {
	case "prod", "product", "release":
		return "release"
	case "test", "testing":
		return "test"
	default:
		return "debug"
	}
}

func setupValidator() {
	if binding.Validator == nil {
		return
	}

	_ = binding.Validator.Engine()
}
