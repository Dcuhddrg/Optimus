package system

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLDAPServiceDatabaseGetLDAPConfigReturnsEmptyWhenMissing(t *testing.T) {
	openPersistenceTestDB(t)

	store := (&LDAPService{})
	cfg, err := store.getLDAPConfig(context.Background())

	require.NoError(t, err)
	require.Equal(t, uint(0), cfg.ID)
}
