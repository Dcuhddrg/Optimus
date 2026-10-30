package response

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessKeepsValidFalsyData(t *testing.T) {
	for _, tt := range []struct {
		name string
		data any
		want string
	}{{"empty", nil, `{"code":0,"data":{},"msg":"success"}`}, {"false", false, `{"code":0,"data":false,"msg":"success"}`}, {"zero", 0, `{"code":0,"data":0,"msg":"success"}`}} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			Success(c, http.StatusOK, tt.data)
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, tt.want, w.Body.String())
		})
	}
}
func TestFailuresKeepHTTPAndDataRules(t *testing.T) {
	for _, status := range []int{401, 403, 422, 500, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			Failure(c, status, "rejected")
			require.Equal(t, status, w.Code)
			data := "{}"
			if status == 401 {
				data = "null"
			}
			require.JSONEq(t, `{"code":7,"data":`+data+`,"msg":"rejected"}`, w.Body.String())
		})
	}
}
func TestProtocolSelectorIsExactAndPrefixAware(t *testing.T) {
	for _, prefix := range []string{"/api/v1", "/custom/v2", "/"} {
		t.Run(prefix, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			g := r.Group(prefix)
			g.Use(EnvelopeForRoutes(g.BasePath(), []Route{{Method: "GET", Path: "/system/audit-logs"}}))
			fail := func(c *gin.Context) { AbortFailure(c, 403, "denied") }
			g.GET("/system/audit-logs", fail)
			g.POST("/system/audit-logs", fail)
			g.GET("/system/audit-logs-extra", fail)
			base := prefix
			if base == "/" {
				base = ""
			}
			for _, test := range []struct {
				method, path string
				envelope     bool
			}{{"GET", "/system/audit-logs", true}, {"POST", "/system/audit-logs", false}, {"GET", "/system/audit-logs-extra", false}} {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(test.method, base+test.path, nil))
				require.Equal(t, 403, w.Code)
				if test.envelope {
					require.JSONEq(t, `{"code":7,"data":{},"msg":"denied"}`, w.Body.String())
				} else {
					require.JSONEq(t, `{"message":"denied"}`, w.Body.String())
				}
			}
		})
	}
}
