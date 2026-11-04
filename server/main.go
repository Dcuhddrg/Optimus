package main

import (
	"optimus/server/core"
	"optimus/server/global"
	"optimus/server/initialize"

	"go.uber.org/zap"
)

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
	initialize.InitializeServices(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)
	if err := initialize.InitializeData(global.OPTIMUS_DB); err != nil {
		global.OPTIMUS_STARTUP_ERR = err
		logger.Error("startup degraded: serving with /health unhealthy", zap.Error(err))
	}
}
