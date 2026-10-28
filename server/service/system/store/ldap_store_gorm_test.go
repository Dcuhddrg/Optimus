package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGormLDAPStoreGetLDAPConfigReturnsEmptyWhenMissing(t *testing.T) {
	db := openPersistenceTestDB(t)

	store := NewGormLDAPStore(db)
	cfg, err := store.GetLDAPConfig(context.Background())

	require.NoError(t, err)
	require.Equal(t, uint(0), cfg.ID)
}
