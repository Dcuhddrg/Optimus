package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBcryptHashAndCheck(t *testing.T) {
	hash, err := BcryptHash("admin-secret")
	require.NoError(t, err)
	require.NotEqual(t, "admin-secret", hash)
	require.True(t, BcryptCheck("admin-secret", hash))
	require.False(t, BcryptCheck("wrong-secret", hash))
}
