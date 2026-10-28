package main

import (
	"optimus/server/core"
	"optimus/server/initialize"
)

func main() {
	initializeSystem()
	core.RunServer(initialize.Routers())
}

func initializeSystem() {
	if err := core.Viper(); err != nil {
		panic(err)
	}
	initialize.OtherInit()
	if err := core.Zap(); err != nil {
		panic(err)
	}
	initialize.Startup()
}
