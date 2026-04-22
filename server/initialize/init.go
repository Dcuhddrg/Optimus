package initialize

import (
	"strings"

	"optimus/server/global"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func OtherInit() {
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
