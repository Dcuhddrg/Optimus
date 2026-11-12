package middleware

import (
	"context"
	"errors"
	"strconv"
	"time"

	"optimus/server/global"
	"optimus/server/model/common/response"
	"optimus/server/service"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtService = service.ServiceGroupApp.SystemServiceGroup.JwtService

// AccountDisabledMessage is the 401 msg for a disabled user's session; the
// frontend shows it on the login page.
const AccountDisabledMessage = "account is disabled"

// AccountExpiredMessage is the 401 msg for an external account past its
// expiry (design 3.6).
const AccountExpiredMessage = "account has expired"

// authStateLoader loads the user's current status and roles; tests replace it.
var authStateLoader = func(ctx context.Context, userID uint) (systemService.UserAuthState, error) {
	return service.ServiceGroupApp.SystemServiceGroup.AuthStateService.LoadUserAuthState(ctx, userID)
}

// JWTAuth authenticates the x-token header or cookie, rejects revoked tokens
// and refreshes tokens inside their buffer window, as GVA.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.GetToken(c)
		if token == "" {
			response.NoAuth("not logged in or invalid access", c)
			c.Abort()
			return
		}
		if jwtService.IsBlacklist(token) {
			response.NoAuth("token has been revoked", c)
			utils.ClearToken(c)
			c.Abort()
			return
		}
		j := utils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.TokenExpired) {
				response.NoAuth("login has expired, please log in again", c)
				utils.ClearToken(c)
				c.Abort()
				return
			}
			response.NoAuth(err.Error(), c)
			utils.ClearToken(c)
			c.Abort()
			return
		}

		// Authorize against the user's current status and roles, not the
		// token's snapshot, so changes apply on the next request (design 3.1).
		state, err := authStateLoader(c.Request.Context(), claims.BaseClaims.ID)
		if err != nil {
			response.FailWithMessage("failed to load user state", c)
			c.Abort()
			return
		}
		if !state.Exists {
			response.NoAuth("account does not exist", c)
			utils.ClearToken(c)
			c.Abort()
			return
		}
		if !state.Enabled() {
			response.NoAuth(AccountDisabledMessage, c)
			utils.ClearToken(c)
			c.Abort()
			return
		}
		if state.Expired(time.Now()) {
			response.NoAuth(AccountExpiredMessage, c)
			utils.ClearToken(c)
			c.Abort()
			return
		}

		c.Set(utils.ClaimsContextKey, claims)
		c.Set(utils.RoleIDsContextKey, state.RoleIDs)
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(global.OPTIMUS_CONFIG.JWT.ExpiresTime)
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(dr))
			newToken, _ := j.CreateTokenByOldToken(token, *claims)
			newClaims, _ := j.ParseToken(newToken)
			c.Header("new-token", newToken)
			c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
			utils.SetToken(c, newToken, int(dr.Seconds()))
		}
		c.Next()

		if newToken, exists := c.Get("new-token"); exists {
			c.Header("new-token", newToken.(string))
		}
		if newExpiresAt, exists := c.Get("new-expires-at"); exists {
			c.Header("new-expires-at", newExpiresAt.(string))
		}
	}
}
