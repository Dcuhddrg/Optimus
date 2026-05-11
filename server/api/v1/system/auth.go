package system

import (
	"errors"
	"net/http"

	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) Login(c *gin.Context) {
	var req systemService.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid login request"})
		return
	}
	if h.service == nil {
		writeServiceUnavailable(c, "auth service is not available")
		return
	}

	result, err := h.service.Login(c.Request.Context(), req, systemService.AuditMeta{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		status := http.StatusUnauthorized
		message := "username or password is incorrect"
		if errors.Is(err, systemService.ErrUserDisabled) {
			status = http.StatusForbidden
			message = "account is disabled"
		}
		c.JSON(status, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) CurrentUser(c *gin.Context) {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "claims are required"})
		return
	}
	customClaims, ok := claims.(*utils.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid claims"})
		return
	}
	if h.service == nil {
		writeServiceUnavailable(c, "auth service is not available")
		return
	}

	result, err := h.service.CurrentUser(c.Request.Context(), customClaims.BaseClaims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load current user"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if ok && h.service != nil {
		if customClaims, ok := claims.(*utils.CustomClaims); ok {
			_ = h.service.Logout(c.Request.Context(), customClaims.BaseClaims.ID, systemService.AuditMeta{
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
