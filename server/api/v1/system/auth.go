package system

import (
	"errors"
	"net/http"
	systemModel "optimus/server/model/system"

	systemService "optimus/server/service/system"
	"optimus/server/utils"
	"optimus/server/utils/response"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) Login(c *gin.Context) {
	var req systemModel.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid login request")
		return
	}
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "auth service is not available")
		return
	}

	result, err := h.service.Login(c.Request.Context(), req, auditMeta(c))
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

func (h *AuthHandler) CurrentUser(c *gin.Context) {
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
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "auth service is not available")
		return
	}

	result, err := h.service.CurrentUser(c.Request.Context(), customClaims.BaseClaims.ID)
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to load current user")
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if ok && h.service != nil {
		if customClaims, ok := claims.(*utils.CustomClaims); ok {
			_ = h.service.Logout(c.Request.Context(), customClaims.BaseClaims.ID, auditMeta(c))
		}
	}
	response.Success(c, http.StatusOK, nil)
}
