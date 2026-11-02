package system

import (
	"errors"
	"net/http"
	systemRequest "optimus/server/model/system/request"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type LDAPApi struct{}

func (h *LDAPApi) LDAPConfig(c *gin.Context) {
	if systemServices.LDAPService == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	cfg, err := systemServices.LDAPService.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load ldap config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *LDAPApi) UpdateLDAPConfig(c *gin.Context) {
	if systemServices.LDAPService == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemRequest.LDAPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ldap config request"})
		return
	}
	cfg, err := systemServices.LDAPService.SaveConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save ldap config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *LDAPApi) TestLDAPConfig(c *gin.Context) {
	if systemServices.LDAPService == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemRequest.LDAPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ldap test request"})
		return
	}
	result, err := systemServices.LDAPService.TestConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
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
