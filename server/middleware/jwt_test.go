package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	systemService "optimus/server/service/system"
	systemUsecase "optimus/server/service/system/usecase"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthRejectsMissingAuthorizationHeader(t *testing.T) {
	engine := jwtTestEngine(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "authorization header is required")
}

func TestJWTAuthRejectsMalformedBearerToken(t *testing.T) {
	engine := jwtTestEngine(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Basic nope")
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "invalid authorization header")
}

func TestJWTAuthStoresClaimsForValidToken(t *testing.T) {
	cfg := testMiddlewareJWTConfig()
	global.OPTIMUS_CONFIG.JWT = cfg
	engine := jwtTestEngine(t)
	token, err := systemService.IssueJWT(systemUsecase.TokenSubject{
		ID:        9,
		UUID:      "uuid-9",
		Username:  "admin",
		RoleCodes: []string{"admin"},
	}, cfg, time.Now())
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "admin")
}

func jwtTestEngine(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	originalConfig := global.OPTIMUS_CONFIG
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
	})

	global.OPTIMUS_CONFIG.JWT = testMiddlewareJWTConfig()

	engine := gin.New()
	engine.GET("/private", JWTAuth(), func(c *gin.Context) {
		claims, ok := c.Get(utils.ClaimsContextKey)
		require.True(t, ok)
		c.JSON(http.StatusOK, gin.H{"username": claims.(*utils.CustomClaims).Username})
	})
	return engine
}

func testMiddlewareJWTConfig() config.JWT {
	return config.JWT{
		SigningKey:  "middleware-test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-test",
	}
}
