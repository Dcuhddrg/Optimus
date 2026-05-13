package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubSyncer struct{}

func (stubSyncer) Sync(_ context.Context, _ SyncAccount) ([]RawAsset, error) {
	return nil, nil
}

func TestSyncerRegistryReturnsRegisteredSyncer(t *testing.T) {
	r := NewSyncerRegistry()
	s := stubSyncer{}
	r.Register("aliyun", s)
	got, err := r.Get("aliyun")
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestSyncerRegistryReturnsErrForUnknownProvider(t *testing.T) {
	r := NewSyncerRegistry()
	_, err := r.Get("aws")
	require.ErrorIs(t, err, ErrProviderNotRegistered)
}
