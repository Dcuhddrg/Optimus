package persistence

import (
	"context"
	"testing"

	"optimus/server/service/cloud/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAssetStoreUpsertCreatesNewAsset(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAssetStore(db)

	result, err := store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)

	page, err := store.Query(context.Background(), usecase.AssetFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "i-001", page.Items[0].ResourceID)
}

func TestGormAssetStoreUpsertUpdatesExistingAsset(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAssetStore(db)

	_, err := store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)

	_, err = store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1-updated", Region: "cn-hangzhou", Status: "Stopped"},
	})
	require.NoError(t, err)

	page, err := store.Query(context.Background(), usecase.AssetFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "web-1-updated", page.Items[0].Name)
	require.Equal(t, "Stopped", page.Items[0].Status)
}

func TestGormAssetStoreQueryFiltersByProvider(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAssetStore(db)

	_, err := store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	_, err = store.Upsert(context.Background(), 2, "aws", []usecase.RawAsset{
		{ResourceType: "ec2_instance", ResourceID: "i-abc", Region: "us-east-1", Status: "running"},
	})
	require.NoError(t, err)

	provider := "aliyun"
	page, err := store.Query(context.Background(), usecase.AssetFilter{Provider: &provider, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "aliyun", page.Items[0].Provider)
}
