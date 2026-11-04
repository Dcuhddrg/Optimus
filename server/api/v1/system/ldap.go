package system

import (
	"errors"
	"optimus/server/global"
	"optimus/server/model/common/response"
	systemRequest "optimus/server/model/system/request"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type LDAPApi struct{}

func (h *LDAPApi) LDAPConfig(c *gin.Context) {
	svc := ldapService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	cfg, err := svc.GetConfig(c.Request.Context())
	if err != nil {
		response.FailWithMessage("failed to load ldap config", c)
		return
	}
	response.OkWithData(cfg, c)
}

func (h *LDAPApi) UpdateLDAPConfig(c *gin.Context) {
	svc := ldapService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemRequest.LDAPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid ldap config request", c)
		return
	}
	cfg, err := svc.SaveConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
	if err != nil {
		response.FailWithMessage("failed to save ldap config", c)
		return
	}
	response.OkWithData(cfg, c)
}

func (h *LDAPApi) TestLDAPConfig(c *gin.Context) {
	svc := ldapService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemRequest.LDAPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid ldap test request", c)
		return
	}
	result, err := svc.TestConfig(c.Request.Context(), actorID(c), req, auditMeta(c))
	if errors.Is(err, systemService.ErrInvalidLDAPConfig) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("ldap connection test failed", c)
		return
	}
	response.OkWithData(result, c)
}
