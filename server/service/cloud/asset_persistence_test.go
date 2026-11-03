package cloud

import (
	"context"
	cloudRequest "optimus/server/model/cloud/request"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetPersistenceUpsertCreatesNewAsset(t *testing.T) {
	accountSvc := newTestAccountService(t)
	store := new(AssetRegistryService)

	account, err := accountSvc.Create(context.Background(), 7, testAccountRow("prod", "aliyun", "123"))
	require.NoError(t, err)

	result, err := store.upsertAssets(context.Background(), account.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)

	page, err := store.QueryAssets(context.Background(), AssetFilter{AccessibleProjectIDs: []uint{7}, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "i-001", page.Items[0].ResourceID)
}

func TestAssetPersistenceUpsertUpdatesExistingAsset(t *testing.T) {
	accountSvc := newTestAccountService(t)
	store := new(AssetRegistryService)

	account, err := accountSvc.Create(context.Background(), 7, testAccountRow("prod", "aliyun", "123"))
	require.NoError(t, err)

	_, err = store.upsertAssets(context.Background(), account.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)

	_, err = store.upsertAssets(context.Background(), account.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1-updated", Region: "cn-hangzhou", Status: "Stopped"},
	})
	require.NoError(t, err)

	page, err := store.QueryAssets(context.Background(), AssetFilter{AccessibleProjectIDs: []uint{7}, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "web-1-updated", page.Items[0].Name)
	require.Equal(t, "Stopped", page.Items[0].Status)
}

func TestAssetPersistenceQueryFiltersByProvider(t *testing.T) {
	accountSvc := newTestAccountService(t)
	store := new(AssetRegistryService)

	aliyunAccount, err := accountSvc.Create(context.Background(), 7, testAccountRow("aliyun", "aliyun", "123"))
	require.NoError(t, err)
	awsAccount, err := accountSvc.Create(context.Background(), 7, testAccountRow("aws", "aws", "456"))
	require.NoError(t, err)

	_, err = store.upsertAssets(context.Background(), aliyunAccount.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	_, err = store.upsertAssets(context.Background(), awsAccount.ID, "aws", []RawAsset{
		{ResourceType: "ec2_instance", ResourceID: "i-abc", Region: "us-east-1", Status: "running"},
	})
	require.NoError(t, err)

	provider := "aliyun"
	page, err := store.QueryAssets(context.Background(), AssetFilter{AccessibleProjectIDs: []uint{7}, Provider: &provider, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "aliyun", page.Items[0].Provider)
}

func TestAssetPersistenceQueryFiltersByAccessibleProjects(t *testing.T) {
	accountSvc := newTestAccountService(t)
	store := new(AssetRegistryService)

	account1, err := accountSvc.Create(context.Background(), 7, testAccountRow("prod", "aliyun", "123"))
	require.NoError(t, err)
	account2, err := accountSvc.Create(context.Background(), 8, testAccountRow("other", "aliyun", "456"))
	require.NoError(t, err)

	_, err = store.upsertAssets(context.Background(), account1.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "visible", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	_, err = store.upsertAssets(context.Background(), account2.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-002", Name: "hidden", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)

	page, err := store.QueryAssets(context.Background(), AssetFilter{AccessibleProjectIDs: []uint{7}, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "visible", page.Items[0].Name)
}

func TestAssetPersistenceQueryFiltersByAccountTag(t *testing.T) {
	accountSvc := newTestAccountService(t)
	assetStore := new(AssetRegistryService)

	prodAccount, err := accountSvc.Create(context.Background(), 1, testAccountRow("prod-acc", "aliyun", "p1"))
	require.NoError(t, err)
	_, err = accountSvc.UpdateTags(context.Background(), []uint{1}, prodAccount.ID, cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "prod"}})
	require.NoError(t, err)

	devAccount, err := accountSvc.Create(context.Background(), 1, testAccountRow("dev-acc", "aliyun", "d1"))
	require.NoError(t, err)
	_, err = accountSvc.UpdateTags(context.Background(), []uint{1}, devAccount.ID, cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "dev"}})
	require.NoError(t, err)

	_, err = assetStore.upsertAssets(context.Background(), prodAccount.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs", ResourceID: "i-1", Name: "prod-vm"},
	})
	require.NoError(t, err)
	_, err = assetStore.upsertAssets(context.Background(), devAccount.ID, "aliyun", []RawAsset{
		{ResourceType: "ecs", ResourceID: "i-2", Name: "dev-vm"},
	})
	require.NoError(t, err)

	page, err := assetStore.QueryAssets(context.Background(), AssetFilter{
		AccessibleProjectIDs: []uint{1},
		Tags:                 []TagFilter{{Key: "environment", Value: "prod", Exists: true}},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "prod-vm", page.Items[0].Name)
}

func testAccountRow(name, provider, accountID string) cloudRequest.CreateAccountRequest {
	return cloudRequest.CreateAccountRequest{
		Name:           name,
		Provider:       provider,
		AccountID:      accountID,
		CredentialType: "ak_sk",
	}
}
