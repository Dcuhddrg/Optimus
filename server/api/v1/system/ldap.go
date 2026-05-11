package system

import (
	"errors"
	"net/http"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

func (h *LDAPHandler) LDAPConfig(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	cfg, err := h.service.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load ldap config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *LDAPHandler) UpdateLDAPConfig(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemService.LDAPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ldap config request"})
		return
	}
	cfg, err := h.service.SaveConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save ldap config"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *LDAPHandler) TestLDAPConfig(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemService.LDAPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid ldap test request"})
		return
	}
	result, err := h.service.TestConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
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
