package cloud

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseTagFiltersPreservesRepeatedQueryValues(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/cloud/accounts?tag=env%3Dprod&tag=team%3Dcore", nil)

	filters, err := parseTagFilters(c)
	require.NoError(t, err)
	require.Len(t, filters, 2)
	require.Equal(t, "env", filters[0].Key)
	require.Equal(t, "prod", filters[0].Value)
	require.Equal(t, "team", filters[1].Key)
	require.Equal(t, "core", filters[1].Value)
}
