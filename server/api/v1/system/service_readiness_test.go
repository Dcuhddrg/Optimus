package system

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"optimus/server/global"
)

func TestDatabaseBackedSystemAPIsAreUnavailableWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = previousDB })
	global.OPTIMUS_DB = nil
	t.Run("auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/system/auth/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		require.NotPanics(t, func() { (&AuthApi{}).Login(c) })
		require.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
	t.Run("ldap", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/system/ldap/config", nil)
		require.NotPanics(t, func() { (&LDAPApi{}).LDAPConfig(c) })
		require.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}
