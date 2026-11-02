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

func TestAssetSearchMapsQueryFieldsAndRepeatedTagsToService(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?page=2&pageSize=50&accountId=7&provider=aliyun&region=cn-test&resourceType=ecs&status=running&keyword=hello&tag=env%3Dprod&tag=owner%3D%2A&tag=__none__%3Aretired", nil)
	f, err := parseAssetFilter(c)
	require.NoError(t, err)
	require.Equal(t, 2, f.Page)
	require.Equal(t, 50, f.PageSize)
	require.NotNil(t, f.AccountID)
	require.Equal(t, uint(7), *f.AccountID)
	require.NotNil(t, f.Provider)
	require.Equal(t, "aliyun", *f.Provider)
	require.NotNil(t, f.Region)
	require.Equal(t, "cn-test", *f.Region)
	require.NotNil(t, f.ResourceType)
	require.Equal(t, "ecs", *f.ResourceType)
	require.NotNil(t, f.Status)
	require.Equal(t, "running", *f.Status)
	require.NotNil(t, f.Keyword)
	require.Equal(t, "hello", *f.Keyword)
	require.Nil(t, f.AccessibleProjectIDs)
	require.Len(t, f.Tags, 3)
	require.Equal(t, "env", f.Tags[0].Key)
	require.Equal(t, "prod", f.Tags[0].Value)
	require.True(t, f.Tags[0].Exists)
	require.Equal(t, "owner", f.Tags[1].Key)
	require.Empty(t, f.Tags[1].Value)
	require.True(t, f.Tags[1].Exists)
	require.Equal(t, "retired", f.Tags[2].Key)
	require.False(t, f.Tags[2].Exists)
	invalid, _ := gin.CreateTestContext(httptest.NewRecorder())
	invalid.Request = httptest.NewRequest("GET", "/?tag=invalid", nil)
	_, err = parseAssetFilter(invalid)
	require.Error(t, err)
}
