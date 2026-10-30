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
		{"POST", "/system/auth/login", "{", 400, `{"code":7,"data":{},"msg":"invalid login request"}`},
		{"GET", "/system/auth/me", "", 401, `{"code":7,"data":null,"msg":"authorization header is required"}`},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, "/api/v1"+tc.path, strings.NewReader(tc.body))
		Routers().ServeHTTP(w, req)
		require.Equal(t, tc.status, w.Code)
		if tc.status == 400 {
			require.JSONEq(t, tc.expected, w.Body.String())
		} else {
			require.Contains(t, w.Body.String(), `"code":7`)
			require.Contains(t, w.Body.String(), `"data":null`)
		}
	}
}
