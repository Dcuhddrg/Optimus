package system

import (
	"errors"
	"net/http"
	systemRequest "optimus/server/model/system/request"

	"optimus/server/model/common/response"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *AuthApi) Login(c *gin.Context) {
	var req systemRequest.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid login request")
		return
	}
	if systemServices.AuthService == nil {
		response.Failure(c, http.StatusServiceUnavailable, "auth service is not available")
		return
	}

	result, err := systemServices.AuthService.Login(c.Request.Context(), req, auditMeta(c))
	if err != nil {
		status := http.StatusUnauthorized
		message := "username or password is incorrect"
		if errors.Is(err, systemService.ErrUserDisabled) {
			status = http.StatusForbidden
			message = "account is disabled"
		}
		response.Failure(c, status, message)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *AuthApi) CurrentUser(c *gin.Context) {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if !ok {
		response.Failure(c, http.StatusUnauthorized, "claims are required")
		return
	}
	customClaims, ok := claims.(*utils.CustomClaims)
	if !ok {
		response.Failure(c, http.StatusUnauthorized, "invalid claims")
		return
	}
	if systemServices.AuthService == nil {
		response.Failure(c, http.StatusServiceUnavailable, "auth service is not available")
		return
	}

	result, err := systemServices.AuthService.CurrentUser(c.Request.Context(), customClaims.BaseClaims.ID)
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to load current user")
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *AuthApi) Logout(c *gin.Context) {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if ok && systemServices.AuthService != nil {
		if customClaims, ok := claims.(*utils.CustomClaims); ok {
			_ = systemServices.AuthService.Logout(c.Request.Context(), customClaims.BaseClaims.ID, auditMeta(c))
		}
	}
	response.Success(c, http.StatusOK, nil)
}
