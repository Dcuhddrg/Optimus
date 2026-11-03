package cloud

import (
	"context"
	cloudRequest "optimus/server/model/cloud/request"
	"testing"

	"github.com/stretchr/testify/require"
)

type fixedSyncer struct {
	assets []RawAsset
}

func (s fixedSyncer) Sync(_ context.Context, _ SyncAccount) ([]RawAsset, error) {
	return s.assets, nil
}

func TestAssetRegistryServiceSyncAccountUpsertsFetchedAssets(t *testing.T) {
	accountSvc := newTestAccountService(t)

	_, err := accountSvc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	}})

	svc := NewAssetRegistryService(accountSvc, registry)
	result, err := svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)
	page, err := svc.QueryAssets(context.Background(), AssetFilter{AccessibleProjectIDs: []uint{7}})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "i-001", page.Items[0].ResourceID)
}

func TestAssetRegistryServiceSyncAccountReturnsErrForUnknownProvider(t *testing.T) {
	accountSvc := newTestAccountService(t)
	_, err := accountSvc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name: "aws-acc", Provider: "aws", AccountID: "456",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	svc := NewAssetRegistryService(accountSvc, NewSyncerRegistry())
	_, err = svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.ErrorIs(t, err, ErrProviderNotRegistered)
}

func TestAssetRegistryServiceSyncAccountSetsActiveStatusOnSuccess(t *testing.T) {
	accountSvc := newTestAccountService(t)

	_, err := accountSvc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{}})

	svc := NewAssetRegistryService(accountSvc, registry)
	_, err = svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.NoError(t, err)

	// Verify sync status was set to active
	account, err := accountSvc.Get(context.Background(), []uint{7}, 1)
	require.NoError(t, err)
	require.Equal(t, "active", account.Status)
	require.NotNil(t, account.LastSyncedAt)
}

func TestAssetRegistryServiceSyncAccountSetsErrorStatusOnFailure(t *testing.T) {
	accountSvc := newTestAccountService(t)

	_, err := accountSvc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name: "prod", Provider: "aws", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	// No syncer registered for "aws" → ErrProviderNotRegistered
	svc := NewAssetRegistryService(accountSvc, NewSyncerRegistry())
	_, err = svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.Error(t, err)

	// Verify sync status was set to error
	account, err := accountSvc.Get(context.Background(), []uint{7}, 1)
	require.NoError(t, err)
	require.Equal(t, "error", account.Status)
}

func TestAssetRegistryServiceSyncAccountRejectsInaccessibleProject(t *testing.T) {
	accountSvc := newTestAccountService(t)

	_, err := accountSvc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{{ResourceID: "i-001"}}})

	svc := NewAssetRegistryService(accountSvc, registry)
	_, err = svc.SyncAccount(context.Background(), []uint{8}, 1)
	require.ErrorIs(t, err, ErrAccountNotFound)
	page, err := svc.QueryAssets(context.Background(), AssetFilter{AccessibleProjectIDs: []uint{7}})
	require.NoError(t, err)
	require.Empty(t, page.Items)
}

func TestAssetRegistryServiceQueryAssetsReturnsEmptyWhenNoAccessibleProjects(t *testing.T) {
	svc := NewAssetRegistryService(nil, NewSyncerRegistry())

	page, err := svc.QueryAssets(context.Background(), AssetFilter{})
	require.NoError(t, err)
	require.Empty(t, page.Items)
	require.Zero(t, page.Total)
}
