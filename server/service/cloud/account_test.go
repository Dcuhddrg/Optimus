package cloud

import (
	"context"
	cloudRequest "optimus/server/model/cloud/request"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestAccountService(t *testing.T) *CloudAccountService {
	t.Helper()
	openCloudTestDB(t)
	svc, err := NewCloudAccountService(make([]byte, 32))
	require.NoError(t, err)
	return svc
}

func sampleCreateAccountRequest() cloudRequest.CreateAccountRequest {
	return cloudRequest.CreateAccountRequest{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "12345678",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	}
}

func TestNewCloudAccountServiceRejectsInvalidKey(t *testing.T) {
	_, err := NewCloudAccountService([]byte("tooshort"))
	require.ErrorIs(t, err, ErrInvalidEncryptionKey)
}

func TestCloudAccountServiceCreateMasksCredential(t *testing.T) {
	svc := newTestAccountService(t)
	view, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
	require.NoError(t, err)
	require.Equal(t, "***", view.Credential)
	require.Equal(t, "prod", view.Name)
	require.Equal(t, "aliyun", view.Provider)
	require.Equal(t, uint(7), view.ProjectID)
}

func TestCloudAccountServiceGetForSyncDecryptsCredential(t *testing.T) {
	svc := newTestAccountService(t)
	_, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
	require.NoError(t, err)

	syncAcc, err := svc.GetForSync(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "ak", syncAcc.Credential["access_key_id"])
	require.Equal(t, "sk", syncAcc.Credential["access_key_secret"])
}

func TestCloudAccountServiceDeleteRemovesAccount(t *testing.T) {
	svc := newTestAccountService(t)
	_, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), []uint{7}, 1))
	_, err = svc.Get(context.Background(), []uint{7}, 1)
	require.ErrorIs(t, err, ErrAccountNotFound)
}

func TestCloudAccountServiceGetNotFoundReturnsErrAccountNotFound(t *testing.T) {
	svc := newTestAccountService(t)
	_, err := svc.Get(context.Background(), []uint{7}, 9999)
	require.ErrorIs(t, err, ErrAccountNotFound)
}

func TestCloudAccountServiceUpdateEmptyCredentialIsNoop(t *testing.T) {
	svc := newTestAccountService(t)
	_, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
	require.NoError(t, err)
	original, err := findAccount(context.Background(), []uint{7}, 1)
	require.NoError(t, err)

	newName := "renamed"
	_, err = svc.Update(context.Background(), []uint{7}, 1, cloudRequest.UpdateAccountRequest{Name: &newName, Credential: map[string]string{}})
	require.NoError(t, err)
	updated, err := findAccount(context.Background(), []uint{7}, 1)
	require.NoError(t, err)
	require.Equal(t, original.CredentialJSON, updated.CredentialJSON)
	require.Equal(t, "renamed", updated.Name)
}

func TestCloudAccountServiceListMasksAllCredentials(t *testing.T) {
	svc := newTestAccountService(t)
	for i := 0; i < 3; i++ {
		_, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
		require.NoError(t, err)
	}
	views, err := svc.List(context.Background(), []uint{7}, nil)
	require.NoError(t, err)
	require.Len(t, views, 3)
	for _, v := range views {
		require.Equal(t, "***", v.Credential)
	}
}

func TestCloudAccountServiceListFiltersByAccessibleProjects(t *testing.T) {
	svc := newTestAccountService(t)
	_, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
	require.NoError(t, err)
	other := sampleCreateAccountRequest()
	other.Name = "other"
	_, err = svc.Create(context.Background(), 8, other)
	require.NoError(t, err)

	views, err := svc.List(context.Background(), []uint{7}, nil)
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, uint(7), views[0].ProjectID)
	require.Equal(t, "prod", views[0].Name)
}

func TestCreateAcceptsTags(t *testing.T) {
	svc := newTestAccountService(t)

	req := sampleCreateAccountRequest()
	req.Tags = map[string]string{"environment": "prod", "team": "sre"}
	view, err := svc.Create(context.Background(), 7, req)
	require.NoError(t, err)
	require.Equal(t, "prod", view.Tags["environment"])
	require.Equal(t, "sre", view.Tags["team"])
}

func TestCreateRejectsInvalidTagKey(t *testing.T) {
	svc, err := NewCloudAccountService(make([]byte, 32))
	require.NoError(t, err)

	req := sampleCreateAccountRequest()
	req.Tags = map[string]string{"Environment": "prod"}
	_, err = svc.Create(context.Background(), 7, req)
	require.ErrorIs(t, err, ErrInvalidTagKey)
}

func TestUpdateTagsFiltersByAccessibleProjects(t *testing.T) {
	svc := newTestAccountService(t)

	view, err := svc.Create(context.Background(), 7, sampleCreateAccountRequest())
	require.NoError(t, err)

	_, err = svc.UpdateTags(context.Background(), []uint{8}, view.ID,
		cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "dev"}})
	require.ErrorIs(t, err, ErrAccountNotFound)

	updated, err := svc.UpdateTags(context.Background(), []uint{7}, view.ID,
		cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "dev"}})
	require.NoError(t, err)
	require.Equal(t, "dev", updated.Tags["environment"])
}

func TestListFiltersByTag(t *testing.T) {
	svc := newTestAccountService(t)

	req1 := sampleCreateAccountRequest()
	req1.Tags = map[string]string{"environment": "prod"}
	_, err := svc.Create(context.Background(), 7, req1)
	require.NoError(t, err)

	req2 := sampleCreateAccountRequest()
	req2.Name = "other"
	req2.Tags = map[string]string{"environment": "dev"}
	_, err = svc.Create(context.Background(), 7, req2)
	require.NoError(t, err)

	views, err := svc.List(context.Background(), []uint{7},
		[]TagFilter{{Key: "environment", Value: "prod", Exists: true}})
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, "prod", views[0].Tags["environment"])
}
