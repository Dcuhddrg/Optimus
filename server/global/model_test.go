package global_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"optimus/server/global"
	"reflect"
	"testing"
	"time"
)

func TestCommonModelRetainsExistingFieldsWithoutSoftDelete(t *testing.T) {
	value := global.OPTIMUS_MODEL{ID: 7, CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), UpdatedAt: time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC)}
	require.Equal(t, 3, reflect.TypeOf(value).NumField())
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":7,"createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-03T03:04:05Z"}`, string(encoded))
}
