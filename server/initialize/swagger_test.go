package initialize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// As GVA, the Swagger UI and spec are public under the router prefix and
// cover every registered route.
func TestSwaggerDocServedWithAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := Routers()

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/swagger/doc.json", nil))
	require.Equal(t, http.StatusOK, recorder.Code)

	var spec struct {
		BasePath string                                `json:"basePath"`
		Paths    map[string]map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &spec))
	require.Equal(t, "/api/v1", spec.BasePath)

	documented := map[string]bool{}
	for path, methods := range spec.Paths {
		for method := range methods {
			documented[method+" "+path] = true
		}
	}
	for _, route := range engine.Routes() {
		if route.Path == "/api/v1/swagger/*any" {
			continue
		}
		path := ginParam.ReplaceAllString(strings.TrimPrefix(route.Path, "/api/v1"), "{$1}")
		require.Truef(t, documented[strings.ToLower(route.Method)+" "+path], "route %s %s has no Swagger annotation", route.Method, route.Path)
	}
}

// ginParam converts gin's :name path parameters to Swagger's {name}.
var ginParam = regexp.MustCompile(`:([A-Za-z]+)`)
