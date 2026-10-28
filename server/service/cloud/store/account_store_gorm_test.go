package store

import (
	"context"
	"testing"

	cloudService "optimus/server/service/cloud"

	"github.com/stretchr/testify/require"
)

func TestGormAccountStoreCreateAndGet(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, cloudService.AccountRow{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
		CredentialJSON: "enc-cred",
	})
	require.NoError(t, err)
	require.NotZero(t, row.ID)
	require.Equal(t, uint(7), row.ProjectID)
	require.Equal(t, "prod", row.Name)
	require.Equal(t, "active", row.Status)

	got, err := store.Get(context.Background(), []uint{7}, row.ID)
	require.NoError(t, err)
	require.Equal(t, row.ID, got.ID)
	require.Equal(t, "enc-cred", got.CredentialJSON)
}

func TestGormAccountStoreGetNotFoundReturnsErr(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	_, err := store.Get(context.Background(), []uint{7}, 9999)
	require.ErrorIs(t, err, cloudService.ErrAccountNotFound)
}

func TestGormAccountStoreUpdateName(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, cloudService.AccountRow{
		Name:           "old-name",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
		CredentialJSON: "enc",
	})
	require.NoError(t, err)

	newName := "new-name"
	updated, err := store.Update(context.Background(), []uint{7}, row.ID, &newName, nil)
	require.NoError(t, err)
	require.Equal(t, "new-name", updated.Name)
}

func TestGormAccountStoreSoftDelete(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, cloudService.AccountRow{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
		CredentialJSON: "enc",
	})
	require.NoError(t, err)

	require.NoError(t, store.Delete(context.Background(), []uint{7}, row.ID))

	_, err = store.Get(context.Background(), []uint{7}, row.ID)
	require.ErrorIs(t, err, cloudService.ErrAccountNotFound)
}

func TestGormAccountStoreGetRejectsAccountInOtherProject(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, cloudService.AccountRow{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
		CredentialJSON: "enc",
	})
	require.NoError(t, err)

	_, err = store.Get(context.Background(), []uint{8}, row.ID)
	require.ErrorIs(t, err, cloudService.ErrAccountNotFound)
}

func TestGormAccountStoreUpdateTagsRoundTrip(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 1, cloudService.AccountRow{
		Name:           "n",
		Provider:       "aliyun",
		AccountID:      "a",
		CredentialType: "ak_sk",
		CredentialJSON: "enc",
	})
	require.NoError(t, err)

	updated, err := store.UpdateTags(context.Background(), []uint{1}, row.ID,
		map[string]string{"environment": "prod", "team": "sre"})
	require.NoError(t, err)
	require.Equal(t, "prod", updated.Tags["environment"])
	require.Equal(t, "sre", updated.Tags["team"])

	got, err := store.Get(context.Background(), []uint{1}, row.ID)
	require.NoError(t, err)
	require.Equal(t, "prod", got.Tags["environment"])
}

func TestGormAccountStoreListTagFilterEnvironmentProd(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	r1, err := store.Create(context.Background(), 1, cloudService.AccountRow{
		Name:           "p1",
		Provider:       "aliyun",
		AccountID:      "a1",
		CredentialType: "ak_sk",
		CredentialJSON: "e1",
	})
	require.NoError(t, err)
	_, err = store.UpdateTags(context.Background(), []uint{1}, r1.ID, map[string]string{"environment": "prod"})
	require.NoError(t, err)

	r2, err := store.Create(context.Background(), 1, cloudService.AccountRow{
		Name:           "p2",
		Provider:       "aliyun",
		AccountID:      "a2",
		CredentialType: "ak_sk",
		CredentialJSON: "e2",
	})
	require.NoError(t, err)
	_, err = store.UpdateTags(context.Background(), []uint{1}, r2.ID, map[string]string{"environment": "dev"})
	require.NoError(t, err)

	rows, err := store.List(context.Background(), []uint{1},
		[]cloudService.TagFilter{{Key: "environment", Value: "prod", Exists: true}})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "p1", rows[0].Name)
}

func TestGormAccountStoreListTagFilterNoneKey(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	r1, err := store.Create(context.Background(), 1, cloudService.AccountRow{
		Name:           "tagged",
		Provider:       "aliyun",
		AccountID:      "a1",
		CredentialType: "ak_sk",
		CredentialJSON: "e1",
	})
	require.NoError(t, err)
	_, err = store.UpdateTags(context.Background(), []uint{1}, r1.ID, map[string]string{"environment": "prod"})
	require.NoError(t, err)

	_, err = store.Create(context.Background(), 1, cloudService.AccountRow{
		Name:           "untagged",
		Provider:       "aliyun",
		AccountID:      "a2",
		CredentialType: "ak_sk",
		CredentialJSON: "e2",
	})
	require.NoError(t, err)

	rows, err := store.List(context.Background(), []uint{1},
		[]cloudService.TagFilter{{Key: "environment", Exists: false}})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "untagged", rows[0].Name)
}
