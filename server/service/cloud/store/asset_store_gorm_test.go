package store

import (
	"context"
	"testing"

	cloudService "optimus/server/service/cloud"

	"github.com/stretchr/testify/require"
)

func TestGormAssetStoreUpsertCreatesNewAsset(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	store := NewGormAssetStore(db)

	account, err := accountStore.Create(context.Background(), 7, testAccountRow("prod", "aliyun", "123"))
	require.NoError(t, err)

	result, err := store.Upsert(context.Background(), account.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)

	page, err := store.Query(context.Background(), cloudService.AssetFilter{AccessibleProjectIDs: []uint{7}, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "i-001", page.Items[0].ResourceID)
}

func TestGormAssetStoreUpsertUpdatesExistingAsset(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	store := NewGormAssetStore(db)

	account, err := accountStore.Create(context.Background(), 7, testAccountRow("prod", "aliyun", "123"))
	require.NoError(t, err)

	_, err = store.Upsert(context.Background(), account.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)

	_, err = store.Upsert(context.Background(), account.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1-updated", Region: "cn-hangzhou", Status: "Stopped"},
	})
	require.NoError(t, err)

	page, err := store.Query(context.Background(), cloudService.AssetFilter{AccessibleProjectIDs: []uint{7}, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "web-1-updated", page.Items[0].Name)
	require.Equal(t, "Stopped", page.Items[0].Status)
}

func TestGormAssetStoreQueryFiltersByProvider(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	store := NewGormAssetStore(db)

	aliyunAccount, err := accountStore.Create(context.Background(), 7, testAccountRow("aliyun", "aliyun", "123"))
	require.NoError(t, err)
	awsAccount, err := accountStore.Create(context.Background(), 7, testAccountRow("aws", "aws", "456"))
	require.NoError(t, err)

	_, err = store.Upsert(context.Background(), aliyunAccount.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	_, err = store.Upsert(context.Background(), awsAccount.ID, "aws", []cloudService.RawAsset{
		{ResourceType: "ec2_instance", ResourceID: "i-abc", Region: "us-east-1", Status: "running"},
	})
	require.NoError(t, err)

	provider := "aliyun"
	page, err := store.Query(context.Background(), cloudService.AssetFilter{AccessibleProjectIDs: []uint{7}, Provider: &provider, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "aliyun", page.Items[0].Provider)
}

func TestGormAssetStoreQueryFiltersByAccessibleProjects(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	store := NewGormAssetStore(db)

	account1, err := accountStore.Create(context.Background(), 7, testAccountRow("prod", "aliyun", "123"))
	require.NoError(t, err)
	account2, err := accountStore.Create(context.Background(), 8, testAccountRow("other", "aliyun", "456"))
	require.NoError(t, err)

	_, err = store.Upsert(context.Background(), account1.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "visible", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	_, err = store.Upsert(context.Background(), account2.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-002", Name: "hidden", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)

	page, err := store.Query(context.Background(), cloudService.AssetFilter{AccessibleProjectIDs: []uint{7}, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "visible", page.Items[0].Name)
}

func TestGormAssetStoreQueryFiltersByAccountTag(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	assetStore := NewGormAssetStore(db)

	prodAccount, err := accountStore.Create(context.Background(), 1, testAccountRow("prod-acc", "aliyun", "p1"))
	require.NoError(t, err)
	_, err = accountStore.UpdateTags(context.Background(), []uint{1}, prodAccount.ID, map[string]string{"environment": "prod"})
	require.NoError(t, err)

	devAccount, err := accountStore.Create(context.Background(), 1, testAccountRow("dev-acc", "aliyun", "d1"))
	require.NoError(t, err)
	_, err = accountStore.UpdateTags(context.Background(), []uint{1}, devAccount.ID, map[string]string{"environment": "dev"})
	require.NoError(t, err)

	_, err = assetStore.Upsert(context.Background(), prodAccount.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs", ResourceID: "i-1", Name: "prod-vm"},
	})
	require.NoError(t, err)
	_, err = assetStore.Upsert(context.Background(), devAccount.ID, "aliyun", []cloudService.RawAsset{
		{ResourceType: "ecs", ResourceID: "i-2", Name: "dev-vm"},
	})
	require.NoError(t, err)

	page, err := assetStore.Query(context.Background(), cloudService.AssetFilter{
		AccessibleProjectIDs: []uint{1},
		Tags:                 []cloudService.TagFilter{{Key: "environment", Value: "prod", Exists: true}},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "prod-vm", page.Items[0].Name)
}

func testAccountRow(name, provider, accountID string) cloudService.AccountRow {
	return cloudService.AccountRow{
		Name:           name,
		Provider:       provider,
		AccountID:      accountID,
		CredentialType: "ak_sk",
		CredentialJSON: "enc",
	}
}
