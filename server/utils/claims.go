package utils

import (
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const ClaimsContextKey = "claims"

// RoleIDsContextKey holds the requester's current role IDs, loaded per request
// by JWTAuth. Authorization uses these, not the token's RoleIDs (display only).
const RoleIDsContextKey = "roleIds"

// TokenHeader is the request header and cookie name carrying the JWT, as in GVA.
const TokenHeader = "x-token"

type BaseClaims struct {
	UUID      string   `json:"uuid"`
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	RoleIDs   []uint   `json:"roleIds"` // snapshot at login, display only; see RoleIDsContextKey
	RoleCodes []string `json:"roleCodes"`
}

type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

func ClearToken(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}
	if net.ParseIP(host) != nil {
		c.SetCookie(TokenHeader, "", -1, "/", "", false, false)
	} else {
		c.SetCookie(TokenHeader, "", -1, "/", host, false, false)
	}
}

func SetToken(c *gin.Context, token string, maxAge int) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}
	if net.ParseIP(host) != nil {
		c.SetCookie(TokenHeader, token, maxAge, "/", "", false, false)
	} else {
		c.SetCookie(TokenHeader, token, maxAge, "/", host, false, false)
	}
}

// GetToken reads the x-token header, falling back to the x-token cookie and
// refreshing that cookie's lifetime, as GVA's utils.GetToken.
func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get(TokenHeader)
	if token == "" {
		j := NewJWT()
		token, _ = c.Cookie(TokenHeader)
		claims, err := j.ParseToken(token)
		if err != nil {
			zap.L().Debug("x-token cookie could not be parsed")
			return token
		}
		SetToken(c, token, int(claims.ExpiresAt.Unix()-time.Now().Unix()))
	}
	return token
}

func GetClaims(c *gin.Context) (*CustomClaims, error) {
	token := GetToken(c)
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		zap.L().Debug("claims could not be parsed from x-token")
	}
	return claims, err
}

func GetUserID(c *gin.Context) uint {
	if claims, exists := c.Get(ClaimsContextKey); !exists {
		if cl, err := GetClaims(c); err != nil {
			return 0
		} else {
			return cl.BaseClaims.ID
		}
	} else {
		waitUse := claims.(*CustomClaims)
		return waitUse.BaseClaims.ID
	}
}
