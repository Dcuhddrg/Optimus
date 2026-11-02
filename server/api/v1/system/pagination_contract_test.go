package system

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuditPaginationPreservesPermissiveQueryParsing(t *testing.T) {
	for _, tt := range []struct {
		query      string
		page, size int
	}{
		{"", 1, 20}, {"page=2&pageSize=200", 2, 200}, {"page=0&pageSize=0", 1, 20},
		{"page=-3&pageSize=-1", 1, 20}, {"page=bad&pageSize=bad", 1, 20},
		{"page=3&pageSize=201", 3, 20}, {"page=999999999999999999999999&pageSize=999999999999999999999999", int(^uint(0) >> 1), 20}, {"page=&pageSize=", 1, 20},
	} {
		t.Run(tt.query, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/?"+tt.query+"&actorUserId=bad&startTime=bad&endTime=bad", nil)
			f := parseAuditFilter(c)
			require.Equal(t, tt.page, f.Page)
			require.Equal(t, tt.size, f.PageSize)
			require.Nil(t, f.ActorUserID)
			require.Nil(t, f.StartTime)
			require.Nil(t, f.EndTime)
		})
	}
}

func TestAuditSearchMapsAllQueryFieldsToService(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?page=2&pageSize=50&actorUserId=7&action=user.update&resourceType=user&result=success&keyword=hello%20world&startTime=2026-01-01T00:00:00Z&endTime=2026-01-02T00:00:00Z", nil)
	f := parseAuditFilter(c)
	require.Equal(t, 2, f.Page)
	require.Equal(t, 50, f.PageSize)
	require.NotNil(t, f.ActorUserID)
	require.Equal(t, uint(7), *f.ActorUserID)
	require.NotNil(t, f.Action)
	require.Equal(t, "user.update", *f.Action)
	require.NotNil(t, f.ResourceType)
	require.Equal(t, "user", *f.ResourceType)
	require.NotNil(t, f.Result)
	require.Equal(t, "success", *f.Result)
	require.NotNil(t, f.Keyword)
	require.Equal(t, "hello world", *f.Keyword)
	require.NotNil(t, f.StartTime)
	require.Equal(t, "2026-01-01T00:00:00Z", f.StartTime.Format("2006-01-02T15:04:05Z07:00"))
	require.NotNil(t, f.EndTime)
	require.Equal(t, "2026-01-02T00:00:00Z", f.EndTime.Format("2006-01-02T15:04:05Z07:00"))
}
