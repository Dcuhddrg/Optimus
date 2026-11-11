package system

import (
	"errors"
	"optimus/server/global"
	"optimus/server/model/common/response"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"

	systemService "optimus/server/service/system"
	ldapUtil "optimus/server/utils/ldap"

	"github.com/gin-gonic/gin"
)

type LDAPApi struct{}

// swag 只能解析本文件导入的包别名；保留此引用，供 Swagger 注释使用 systemResponse。
var _ = systemResponse.LDAPConfigResult{}

// @Tags LDAP
// @Summary 获取 LDAP 配置
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=systemResponse.LDAPConfigResult} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/ldap/config [get]
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

// @Tags LDAP
// @Summary 保存 LDAP 配置
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body systemRequest.LDAPConfigRequest true "请求体"
// @Success 200 {object} response.Response{data=systemResponse.LDAPConfigResult} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/ldap/config [put]
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
		var locked systemService.IdentityAttrLockedError
		switch {
		case errors.As(err, &locked), errors.Is(err, systemService.ErrInvalidIdentityAttr), errors.Is(err, ldapUtil.ErrInvalidLoginFilter):
			response.FailWithMessage(err.Error(), c)
		default:
			response.FailWithMessage("failed to save ldap config", c)
		}
		return
	}
	response.OkWithData(cfg, c)
}

// @Tags LDAP
// @Summary 测试 LDAP 连接
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body systemRequest.LDAPTestRequest true "请求体"
// @Success 200 {object} response.Response{data=systemResponse.LDAPTestResult} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/ldap/test [post]
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
	if errors.Is(err, systemService.ErrInvalidLDAPConfig) || errors.Is(err, systemService.ErrInvalidIdentityAttr) || errors.Is(err, ldapUtil.ErrInvalidLoginFilter) ||
		errors.Is(err, ldapUtil.ErrNoUsersInScope) || errors.Is(err, ldapUtil.ErrSampleMissingIdentity) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("ldap connection test failed", c)
		return
	}
	response.OkWithData(result, c)
}
