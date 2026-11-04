package initialize

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthProtocolErrors(t *testing.T) {
	for _, tc := range []struct {
		method, path, body string
		status             int
		expected           string
	}{
		{"POST", "/system/auth/login", "{", 200, `{"code":7,"data":{},"msg":"invalid login request"}`},
		{"GET", "/system/auth/me", "", 401, `{"code":7,"data":null,"msg":"not logged in or invalid access"}`},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, "/api/v1"+tc.path, strings.NewReader(tc.body))
		Routers().ServeHTTP(w, req)
		require.Equal(t, tc.status, w.Code)
		require.JSONEq(t, tc.expected, w.Body.String())
	}
}
