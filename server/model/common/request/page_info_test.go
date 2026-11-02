package request

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPageInfoNormalizePreservesPaginationBounds(t *testing.T) {
	for _, tt := range []struct {
		name                           string
		page, size, wantPage, wantSize int
	}{
		{"defaults", 0, 0, 1, 20}, {"negative", -1, -1, 1, 20}, {"upper bound", 2, 200, 2, 200},
		{"oversized resets", 3, 201, 3, 20}, {"valid", 4, 1, 4, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := PageInfo{Page: tt.page, PageSize: tt.size}
			p.Normalize()
			require.Equal(t, tt.wantPage, p.Page)
			require.Equal(t, tt.wantSize, p.PageSize)
		})
	}
}
