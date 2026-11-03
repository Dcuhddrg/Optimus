package system

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"optimus/server/global"
	"optimus/server/service"
	systemService "optimus/server/service/system"
)

func TestStatefulServicesRequireInitializationEvenWhenDatabaseExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB, previousGroup := global.OPTIMUS_DB, service.ServiceGroupApp.SystemServiceGroup
	t.Cleanup(func() {
		global.OPTIMUS_DB = previousDB
		service.ServiceGroupApp.SystemServiceGroup = previousGroup
		RefreshServiceAliases()
	})
	global.OPTIMUS_DB = &gorm.DB{}
	service.ServiceGroupApp.SystemServiceGroup.AuthService = systemService.AuthService{}
	service.ServiceGroupApp.SystemServiceGroup.LDAPService = systemService.LDAPService{}
	RefreshServiceAliases()
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

func TestConstructedServicesPreserveNilDependencySemantics(t *testing.T) {
	require.True(t, systemService.NewAuthService(nil, nil).Ready())
	require.True(t, systemService.NewLDAPService(nil).Ready())
}
