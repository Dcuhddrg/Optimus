package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeAccountStore struct {
	rows map[uint]AccountRow
	next uint
}

func (f *fakeAccountStore) Create(_ context.Context, name, provider, accountID, credentialType, encrypted string) (AccountRow, error) {
	f.next++
	row := AccountRow{ID: f.next, Name: name, Provider: provider, AccountID: accountID, CredentialType: credentialType, CredentialJSON: encrypted, Status: "active"}
	f.rows[f.next] = row
	return row, nil
}

func (f *fakeAccountStore) Get(_ context.Context, id uint) (AccountRow, error) {
	row, ok := f.rows[id]
	if !ok {
		return AccountRow{}, ErrAccountNotFound
	}
	return row, nil
}

func (f *fakeAccountStore) List(_ context.Context) ([]AccountRow, error) {
	rows := make([]AccountRow, 0, len(f.rows))
	for _, r := range f.rows {
		rows = append(rows, r)
	}
	return rows, nil
}

func (f *fakeAccountStore) Update(_ context.Context, id uint, name *string, encryptedCredential *string) (AccountRow, error) {
	row, ok := f.rows[id]
	if !ok {
		return AccountRow{}, ErrAccountNotFound
	}
	if name != nil {
		row.Name = *name
	}
	if encryptedCredential != nil {
		row.CredentialJSON = *encryptedCredential
	}
	f.rows[id] = row
	return row, nil
}

func (f *fakeAccountStore) Delete(_ context.Context, id uint) error {
	if _, ok := f.rows[id]; !ok {
		return ErrAccountNotFound
	}
	delete(f.rows, id)
	return nil
}

func (f *fakeAccountStore) UpdateSyncStatus(_ context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	row, ok := f.rows[id]
	if !ok {
		return ErrAccountNotFound
	}
	row.Status = status
	row.LastSyncedAt = lastSyncedAt
	f.rows[id] = row
	return nil
}

func newTestAccountService() (*CloudAccountService, *fakeAccountStore) {
	store := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	svc, err := NewCloudAccountService(store, key)
	if err != nil {
		panic(err)
	}
	return svc, store
}

func TestNewCloudAccountServiceRejectsInvalidKey(t *testing.T) {
	_, err := NewCloudAccountService(&fakeAccountStore{rows: make(map[uint]AccountRow)}, []byte("tooshort"))
	require.ErrorIs(t, err, ErrInvalidEncryptionKey)
}

func TestCloudAccountServiceCreateMasksCredential(t *testing.T) {
	svc, _ := newTestAccountService()
	view, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "12345678",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)
	require.Equal(t, "***", view.Credential)
	require.Equal(t, "prod", view.Name)
	require.Equal(t, "aliyun", view.Provider)
}

func TestCloudAccountServiceGetForSyncDecryptsCredential(t *testing.T) {
	svc, _ := newTestAccountService()
	_, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "AK", "access_key_secret": "SK"},
	})
	require.NoError(t, err)

	syncAcc, err := svc.GetForSync(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "AK", syncAcc.Credential["access_key_id"])
	require.Equal(t, "SK", syncAcc.Credential["access_key_secret"])
}

func TestCloudAccountServiceDeleteRemovesAccount(t *testing.T) {
	svc, store := newTestAccountService()
	_, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), 1))
	require.Empty(t, store.rows)
}

func TestCloudAccountServiceGetNotFoundReturnsErrAccountNotFound(t *testing.T) {
	svc, _ := newTestAccountService()
	_, err := svc.Get(context.Background(), 9999)
	require.ErrorIs(t, err, ErrAccountNotFound)
}

func TestCloudAccountServiceUpdateEmptyCredentialIsNoop(t *testing.T) {
	svc, store := newTestAccountService()
	_, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "AK"},
	})
	require.NoError(t, err)
	originalCred := store.rows[1].CredentialJSON

	newName := "renamed"
	_, err = svc.Update(context.Background(), 1, UpdateAccountRequest{Name: &newName, Credential: map[string]string{}})
	require.NoError(t, err)
	// credential must not have changed
	require.Equal(t, originalCred, store.rows[1].CredentialJSON)
	require.Equal(t, "renamed", store.rows[1].Name)
}

func TestCloudAccountServiceListMasksAllCredentials(t *testing.T) {
	svc, _ := newTestAccountService()
	for i := 0; i < 3; i++ {
		_, err := svc.Create(context.Background(), CreateAccountRequest{
			Name: "acc", Provider: "aliyun", AccountID: "id",
			CredentialType: "ak_sk",
			Credential:     map[string]string{"access_key_id": "AK"},
		})
		require.NoError(t, err)
	}
	views, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Len(t, views, 3)
	for _, v := range views {
		require.Equal(t, "***", v.Credential)
	}
}
