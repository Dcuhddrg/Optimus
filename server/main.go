package main

import (
	"optimus/server/core"
	"optimus/server/global"
	"optimus/server/initialize"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

func main() {
	initializeSystem()
	core.RunServer()
}

func initializeSystem() {
	global.OPTIMUS_VP = core.Viper()
	initialize.OtherInit()
	logger := core.Zap()
	global.OPTIMUS_LOG = logger
	zap.ReplaceGlobals(logger)
	global.OPTIMUS_DB = initialize.Gorm()
	initialize.Cloud()
	if err := initialize.InitializeData(global.OPTIMUS_DB); err != nil {
		global.OPTIMUS_STARTUP_ERR = err
		logger.Error("startup degraded: serving with /health unhealthy", zap.Error(err))
	}
}
