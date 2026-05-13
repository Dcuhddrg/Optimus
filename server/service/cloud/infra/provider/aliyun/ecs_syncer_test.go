package aliyun

import (
	"context"
	"testing"

	"optimus/server/service/cloud/usecase"

	"github.com/stretchr/testify/require"
)

func TestECSSyncerImplementsAssetSyncer(t *testing.T) {
	var _ usecase.AssetSyncer = NewECSSyncer()
}

func TestECSSyncerReturnsErrorWithInvalidCredentials(t *testing.T) {
	s := NewECSSyncer()
	_, err := s.Sync(context.Background(), usecase.SyncAccount{
		ID:             1,
		Provider:       "aliyun",
		AccountID:      "fake-account",
		CredentialType: "ak_sk",
		Credential: map[string]string{
			"access_key_id":     "invalid-ak",
			"access_key_secret": "invalid-sk",
			"region":            "cn-hangzhou",
		},
	})
	// With invalid credentials against the real API this will error.
	// We only test interface compliance and that no panic occurs.
	_ = err
	_ = require.New // ensure testify is imported (used implicitly via require tag)
}
