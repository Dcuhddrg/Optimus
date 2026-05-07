package system

import (
	"errors"
	"net/http"

	"optimus/server/global"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (a *Api) Login(c *gin.Context) {
	var req systemService.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid login request"})
		return
	}
	if global.OPTIMUS_DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return
	}

	auth := systemService.NewAuthServiceWithLDAP(
		systemService.NewGormAuthStore(global.OPTIMUS_DB),
		global.OPTIMUS_CONFIG.JWT,
		systemService.NewGoLDAPClient(),
	)
	result, err := auth.Login(c.Request.Context(), req, systemService.AuditMeta{
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

func (a *Api) CurrentUser(c *gin.Context) {
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
	if global.OPTIMUS_DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return
	}

	auth := systemService.NewAuthService(systemService.NewGormAuthStore(global.OPTIMUS_DB), global.OPTIMUS_CONFIG.JWT)
	result, err := auth.CurrentUser(c.Request.Context(), customClaims.BaseClaims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load current user"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (a *Api) Logout(c *gin.Context) {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if ok && global.OPTIMUS_DB != nil {
		if customClaims, ok := claims.(*utils.CustomClaims); ok {
			auth := systemService.NewAuthService(systemService.NewGormAuthStore(global.OPTIMUS_DB), global.OPTIMUS_CONFIG.JWT)
			_ = auth.Logout(c.Request.Context(), customClaims.BaseClaims.ID, systemService.AuditMeta{
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
