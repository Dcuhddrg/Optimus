package initialize

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRoutesUnauthorizedUseEnvelope(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{"GET", "/system/users"},
		{"PUT", "/system/users/1"},
		{"PATCH", "/system/users/1/status"},
		{"PUT", "/system/users/1/roles"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			Routers().ServeHTTP(w, httptest.NewRequest(tc.method, "/api/v1"+tc.path, nil))
			require.Equal(t, 401, w.Code)
			require.JSONEq(t, `{"code":7,"data":null,"msg":"authorization header is required"}`, w.Body.String())
		})
	}
}
