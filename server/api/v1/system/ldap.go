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

// @Tags LDAP
// @Summary 预览 LDAP 全量同步
// @Description 只读：分页读取同步范围，与 LDAP 账号比对，返回分类列表和预览 ID。预览 ID 30 分钟内有效，只能由发起人执行一次。
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=systemResponse.LDAPSyncPreview} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/ldap/sync/preview [post]
func (h *LDAPApi) PreviewLDAPSync(c *gin.Context) {
	svc := ldapService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	preview, err := svc.PreviewSync(c.Request.Context(), actorID(c), auditMeta(c))
	if errors.Is(err, systemService.ErrLDAPDisabled) || errors.Is(err, systemService.ErrSyncScopeEmpty) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("ldap sync preview failed; nothing was changed", c)
		return
	}
	response.OkWithData(preview, c)
}

// @Tags LDAP
// @Summary 执行 LDAP 全量同步
// @Description 按预览执行：逐人加锁重读核对后写入；停用、绑定、"不是同一个人"只处理出现在预览里且核对一致的账号。重读出错时停止，结果为 partial。
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body systemRequest.LDAPSyncExecuteRequest true "请求体"
// @Success 200 {object} response.Response{data=systemResponse.LDAPSyncResult} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/ldap/sync/execute [post]
func (h *LDAPApi) ExecuteLDAPSync(c *gin.Context) {
	svc := ldapService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "ldap service is not available")
		return
	}
	var req systemRequest.LDAPSyncExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid ldap sync request", c)
		return
	}
	result, err := svc.ExecuteSync(c.Request.Context(), actorID(c), req, auditMeta(c))
	if errors.Is(err, systemService.ErrSyncPreviewInvalid) || errors.Is(err, systemService.ErrSyncScopeEmpty) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("ldap sync failed; nothing was changed", c)
		return
	}
	response.OkWithData(result, c)
}
