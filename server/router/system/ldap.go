package system

import "github.com/gin-gonic/gin"

type LDAPRouter struct{}

func (r *LDAPRouter) InitLDAPRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/system/ldap/config", ldapApi.LDAPConfig)
	privateGroup.PUT("/system/ldap/config", ldapApi.UpdateLDAPConfig)
	privateGroup.POST("/system/ldap/test", ldapApi.TestLDAPConfig)
}
