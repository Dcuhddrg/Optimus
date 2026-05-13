package cloud

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCloudHandlers(CloudHandlerDeps{})

	engine := gin.New()
	engine.GET("/cloud/accounts", h.Account.ListAccounts)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/accounts", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestAssetHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCloudHandlers(CloudHandlerDeps{})

	engine := gin.New()
	engine.GET("/cloud/assets", h.Asset.ListAssets)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/assets", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}
