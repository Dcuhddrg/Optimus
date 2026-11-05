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

// swag init 只从入口文件解析通用信息和 @Tag 排序，默认 main.go。
// @Tag.Name        Base
// @Tag.Description 探活与健康检查
// @Tag.Name        Auth
// @Tag.Description 登录、当前用户与登出
// @Tag.Name        User
// @Tag.Description 用户管理
// @Tag.Name        Role
// @Tag.Description 角色与权限
// @Tag.Name        LDAP
// @Tag.Description LDAP 配置
// @Tag.Name        Audit
// @Tag.Description 审计日志
// @Tag.Name        Project
// @Tag.Description 项目管理
// @Tag.Name        ProjectMember
// @Tag.Description 项目成员
// @Tag.Name        CloudAccount
// @Tag.Description 云账号
// @Tag.Name        CloudAsset
// @Tag.Description 云资产

// @title                       Optimus Swagger API 接口文档
// @version                     v1
// @description                 Optimus 运维平台后端接口。业务结果均为 HTTP 200 + {code,data,msg}（code 0 成功、7 失败），仅未认证返回 HTTP 401。
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /
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
	initialize.Timer()
	initialize.Cloud()
	if err := initialize.InitializeData(global.OPTIMUS_DB); err != nil {
		global.OPTIMUS_STARTUP_ERR = err
		logger.Error("startup degraded: serving with /health unhealthy", zap.Error(err))
	}
}
