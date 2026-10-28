package cloud

import (
	"context"
	cloudModel "optimus/server/model/cloud"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeAssetStore struct {
	upserted   []RawAsset
	lastFilter AssetFilter
	queried    bool
}

func (f *fakeAssetStore) Upsert(_ context.Context, accountID uint, provider string, assets []RawAsset) (cloudModel.SyncResult, error) {
	f.upserted = append(f.upserted, assets...)
	return cloudModel.SyncResult{Upserted: len(assets)}, nil
}

func (f *fakeAssetStore) Query(_ context.Context, filter AssetFilter) (cloudModel.AssetPage, error) {
	f.queried = true
	f.lastFilter = filter
	return cloudModel.AssetPage{Items: []cloudModel.AssetView{}, Total: 0, Page: filter.Page, PageSize: filter.PageSize}, nil
}

type fixedSyncer struct {
	assets []RawAsset
}

func (s fixedSyncer) Sync(_ context.Context, _ SyncAccount) ([]RawAsset, error) {
	return s.assets, nil
}

func TestAssetRegistryServiceSyncAccountUpsertsFetchedAssets(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc, err := NewCloudAccountService(accountStore, key)
	require.NoError(t, err)

	_, err = accountSvc.Create(context.Background(), 7, cloudModel.CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	assetStore := &fakeAssetStore{}
	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	}})

	svc := NewAssetRegistryService(accountSvc, assetStore, registry)
	result, err := svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)
	require.Equal(t, 1, len(assetStore.upserted))
	require.Equal(t, "i-001", assetStore.upserted[0].ResourceID)
}

func TestAssetRegistryServiceSyncAccountReturnsErrForUnknownProvider(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc, err := NewCloudAccountService(accountStore, key)
	require.NoError(t, err)
	_, err = accountSvc.Create(context.Background(), 7, cloudModel.CreateAccountRequest{
		Name: "aws-acc", Provider: "aws", AccountID: "456",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	svc := NewAssetRegistryService(accountSvc, &fakeAssetStore{}, NewSyncerRegistry())
	_, err = svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.ErrorIs(t, err, ErrProviderNotRegistered)
}

func TestAssetRegistryServiceSyncAccountSetsActiveStatusOnSuccess(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc, err := NewCloudAccountService(accountStore, key)
	require.NoError(t, err)

	_, err = accountSvc.Create(context.Background(), 7, cloudModel.CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{}})

	svc := NewAssetRegistryService(accountSvc, &fakeAssetStore{}, registry)
	_, err = svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.NoError(t, err)

	// Verify sync status was set to active
	require.Equal(t, "active", accountStore.rows[1].Status)
	require.NotNil(t, accountStore.rows[1].LastSyncedAt)
}

func TestAssetRegistryServiceSyncAccountSetsErrorStatusOnFailure(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc, err := NewCloudAccountService(accountStore, key)
	require.NoError(t, err)

	_, err = accountSvc.Create(context.Background(), 7, cloudModel.CreateAccountRequest{
		Name: "prod", Provider: "aws", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	// No syncer registered for "aws" → ErrProviderNotRegistered
	svc := NewAssetRegistryService(accountSvc, &fakeAssetStore{}, NewSyncerRegistry())
	_, err = svc.SyncAccount(context.Background(), []uint{7}, 1)
	require.Error(t, err)

	// Verify sync status was set to error
	require.Equal(t, "error", accountStore.rows[1].Status)
}

func TestAssetRegistryServiceSyncAccountRejectsInaccessibleProject(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc, err := NewCloudAccountService(accountStore, key)
	require.NoError(t, err)

	_, err = accountSvc.Create(context.Background(), 7, cloudModel.CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	assetStore := &fakeAssetStore{}
	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{{ResourceID: "i-001"}}})

	svc := NewAssetRegistryService(accountSvc, assetStore, registry)
	_, err = svc.SyncAccount(context.Background(), []uint{8}, 1)
	require.ErrorIs(t, err, ErrAccountNotFound)
	require.Empty(t, assetStore.upserted)
}

func TestAssetRegistryServiceQueryAssetsReturnsEmptyWhenNoAccessibleProjects(t *testing.T) {
	svc := NewAssetRegistryService(nil, &fakeAssetStore{}, NewSyncerRegistry())

	page, err := svc.QueryAssets(context.Background(), AssetFilter{})
	require.NoError(t, err)
	require.Empty(t, page.Items)
	require.Zero(t, page.Total)
}
