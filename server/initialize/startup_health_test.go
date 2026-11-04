package initialize

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/global"

	"github.com/stretchr/testify/require"
)

func TestInitializeDataReturnsErrorWithoutDatabase(t *testing.T) {
	require.Error(t, InitializeData(nil))
}

func TestHealthReportsStartupFailureAsUnavailable(t *testing.T) {
	savedErr := global.OPTIMUS_STARTUP_ERR
	t.Cleanup(func() { global.OPTIMUS_STARTUP_ERR = savedErr })

	request := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		Routers().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil))
		return w
	}

	global.OPTIMUS_STARTUP_ERR = nil
	w := request("/health")
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"status":"ok"}`, w.Body.String())

	global.OPTIMUS_STARTUP_ERR = errors.New("dial tcp 127.0.0.1:3306: connect: connection refused")
	w = request("/health")
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	require.JSONEq(t, `{"status":"unhealthy"}`, w.Body.String(), "error details must stay in logs, not the public probe")

	for _, path := range []string{"/ping", "/base/ping"} {
		require.Equal(t, http.StatusOK, request(path).Code, path+" stays a liveness probe")
	}
}
