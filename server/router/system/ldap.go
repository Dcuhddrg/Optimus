package system

import "github.com/gin-gonic/gin"

type LDAPRouter struct{}

func (r *LDAPRouter) InitLDAPRouter(privateGroup *gin.RouterGroup) {
	ldapRouter := privateGroup.Group("system/ldap")
	ldapRouter.GET("config", ldapApi.LDAPConfig)
	ldapRouter.PUT("config", ldapApi.UpdateLDAPConfig)
	ldapRouter.POST("test", ldapApi.TestLDAPConfig)
	ldapRouter.POST("sync/preview", ldapApi.PreviewLDAPSync)
	ldapRouter.POST("sync/execute", ldapApi.ExecuteLDAPSync)
}
