package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) LDAPConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": false, "hasBindPassword": false})
}

func (a *Api) UpdateLDAPConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ldap config updated"})
}

func (a *Api) TestLDAPConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "LDAP connection test placeholder"})
}
