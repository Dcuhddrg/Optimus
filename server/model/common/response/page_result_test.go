package response

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageResultPreservesPaginationJSON(t *testing.T) {
	for _, tt := range []struct {
		name string
		list []string
		want string
	}{
		{"empty array", []string{}, `{"list":[],"total":0,"page":1,"pageSize":20}`},
		{"nil list", nil, `{"list":null,"total":0,"page":1,"pageSize":20}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(PageResult[string]{List: tt.list, Page: 1, PageSize: 20})
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(b))
		})
	}
}
