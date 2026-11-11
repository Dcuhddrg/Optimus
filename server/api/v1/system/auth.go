package system

import (
	"errors"
	systemResponse "optimus/server/model/system/response"
	"time"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	"optimus/server/model/common/response"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthApi struct{}

// swag 只能解析本文件导入的包别名；保留此引用，供 Swagger 注释使用 systemResponse。
var _ = systemResponse.LoginResult{}

// @Tags Auth
// @Summary 用户名密码或 LDAP 登录，返回令牌并写入 x-token cookie
// @Accept application/json
// @Produce application/json
// @Param data body systemRequest.LoginRequest true "请求体"
// @Success 200 {object} response.Response{data=systemResponse.LoginResult} "code 0 成功，code 7 业务失败"
// @Router /system/auth/login [post]
func (h *AuthApi) Login(c *gin.Context) {
	svc := authService
	var req systemRequest.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid login request", c)
		return
	}
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("auth service is not available", c)
		return
	}

	result, err := svc.Login(c.Request.Context(), req, auditMeta(c))
	if err != nil {
		message := "username or password is incorrect"
		if errors.Is(err, systemService.ErrUserDisabled) {
			message = "account is disabled"
		}
		for _, refusal := range []error{systemService.ErrLDAPAccountConflict, systemService.ErrLDAPBindingPending, systemService.ErrLDAPIdentityMissing,
			systemService.ErrLDAPIdentityAnomaly, systemService.ErrLDAPDuplicateUsername, systemService.ErrLDAPDirectoryChanging} {
			if errors.Is(err, refusal) {
				message = refusal.Error()
			}
		}
		response.FailWithMessage(message, c)
		return
	}

	// As GVA's TokenNext: the token is returned and also written to the x-token cookie.
	if claims, err := utils.NewJWT().ParseToken(result.Token); err == nil {
		utils.SetToken(c, result.Token, int(time.Until(claims.ExpiresAt.Time).Seconds()))
	}
	response.OkWithData(result, c)
}

// @Tags Auth
// @Summary 获取当前用户、角色与权限
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=systemResponse.CurrentUserResult} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/auth/me [get]
func (h *AuthApi) CurrentUser(c *gin.Context) {
	svc := authService
	claims, ok := c.Get(utils.ClaimsContextKey)
	if !ok {
		response.NoAuth("claims are required", c)
		return
	}
	customClaims, ok := claims.(*utils.CustomClaims)
	if !ok {
		response.NoAuth("invalid claims", c)
		return
	}
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("auth service is not available", c)
		return
	}

	result, err := svc.CurrentUser(c.Request.Context(), customClaims.BaseClaims.ID)
	if err != nil {
		response.FailWithMessage("failed to load current user", c)
		return
	}

	response.OkWithData(result, c)
}

// Logout revokes the current token (GVA's JsonInBlacklist), clears the
// x-token cookie and records the logout audit.
// @Tags Auth
// @Summary 登出：令牌加入黑名单并清除 cookie
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/auth/logout [post]
func (h *AuthApi) Logout(c *gin.Context) {
	svc := authService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("auth service is not available", c)
		return
	}
	token := utils.GetToken(c)
	if err := jwtService.JsonInBlacklist(c.Request.Context(), systemModel.JwtBlacklist{Jwt: token}); err != nil {
		zap.L().Error("revoke jwt failed", zap.Error(err))
		response.FailWithMessage("failed to revoke token", c)
		return
	}
	utils.ClearToken(c)
	if claims, ok := c.Get(utils.ClaimsContextKey); ok {
		if customClaims, ok := claims.(*utils.CustomClaims); ok {
			_ = svc.Logout(c.Request.Context(), customClaims.BaseClaims.ID, auditMeta(c))
		}
	}
	response.Ok(c)
}
