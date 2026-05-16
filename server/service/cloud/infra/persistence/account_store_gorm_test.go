package persistence

import (
	"context"
	"testing"

	"optimus/server/service/cloud/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAccountStoreCreateAndGet(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, "prod", "aliyun", "123", "ak_sk", "enc-cred")
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
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)
}

func TestGormAccountStoreUpdateName(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, "old-name", "aliyun", "123", "ak_sk", "enc")
	require.NoError(t, err)

	newName := "new-name"
	updated, err := store.Update(context.Background(), []uint{7}, row.ID, &newName, nil)
	require.NoError(t, err)
	require.Equal(t, "new-name", updated.Name)
}

func TestGormAccountStoreSoftDelete(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, "prod", "aliyun", "123", "ak_sk", "enc")
	require.NoError(t, err)

	require.NoError(t, store.Delete(context.Background(), []uint{7}, row.ID))

	_, err = store.Get(context.Background(), []uint{7}, row.ID)
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)
}

func TestGormAccountStoreGetRejectsAccountInOtherProject(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 7, "prod", "aliyun", "123", "ak_sk", "enc")
	require.NoError(t, err)

	_, err = store.Get(context.Background(), []uint{8}, row.ID)
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)
}
