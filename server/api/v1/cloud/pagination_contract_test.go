package cloud

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAssetPaginationPreservesPermissiveQueryParsing(t *testing.T) {
	for _, tt := range []struct {
		query      string
		page, size int
	}{
		{"", 1, 20}, {"page=2&pageSize=200", 2, 200}, {"page=0&pageSize=0", 1, 20},
		{"page=-3&pageSize=-1", 1, 20}, {"page=bad&pageSize=bad", 1, 20},
		{"page=3&pageSize=201", 3, 20}, {"page=999999999999999999999999&pageSize=999999999999999999999999", 1, 20}, {"page=&pageSize=", 1, 20},
	} {
		t.Run(tt.query, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/?"+tt.query+"&accountId=bad", nil)
			f, err := parseAssetFilter(c)
			require.NoError(t, err)
			require.Equal(t, tt.page, f.Page)
			require.Equal(t, tt.size, f.PageSize)
			require.Nil(t, f.AccountID)
		})
	}
}
