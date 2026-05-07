package system

import (
	"errors"
	"net/http"

	"optimus/server/global"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

func (a *Api) LDAPConfig(c *gin.Context) {
	service, ok := newLDAPService(c)
	if !ok {
		return
	}
	cfg, err := service.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load ldap config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (a *Api) UpdateLDAPConfig(c *gin.Context) {
	service, ok := newLDAPService(c)
	if !ok {
		return
	}
	var req systemService.LDAPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ldap config request"})
		return
	}
	cfg, err := service.SaveConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save ldap config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (a *Api) TestLDAPConfig(c *gin.Context) {
	service, ok := newLDAPService(c)
	if !ok {
		return
	}
	var req systemService.LDAPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ldap test request"})
		return
	}
	result, err := service.TestConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
	if errors.Is(err, systemService.ErrInvalidLDAPConfig) {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "ldap connection test failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func newLDAPService(c *gin.Context) (*systemService.LDAPService, bool) {
	if global.OPTIMUS_DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return nil, false
	}
	return systemService.NewLDAPService(
		systemService.NewGormLDAPStore(global.OPTIMUS_DB),
		systemService.NewGoLDAPClient(),
	), true
}
