package system

import (
	"errors"
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
		response.FailWithMessage(message, c)
		return
	}

	// As GVA's TokenNext: the token is returned and also written to the x-token cookie.
	if claims, err := utils.NewJWT().ParseToken(result.Token); err == nil {
		utils.SetToken(c, result.Token, int(time.Until(claims.ExpiresAt.Time).Seconds()))
	}
	response.OkWithData(result, c)
}

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
