package cloud

import (
	"context"
	"optimus/server/global"
	cloudRequest "optimus/server/model/cloud/request"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountPersistenceCreateAndGet(t *testing.T) {
	svc := newTestAccountService(t)

	row, err := svc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)
	require.NotZero(t, row.ID)
	require.Equal(t, uint(7), row.ProjectID)
	require.Equal(t, "prod", row.Name)
	require.Equal(t, "active", row.Status)

	got, err := svc.Get(context.Background(), []uint{7}, row.ID)
	require.NoError(t, err)
	require.Equal(t, row.ID, got.ID)
	require.Equal(t, "***", got.Credential)
}

func TestAccountPersistenceGetNotFoundReturnsErr(t *testing.T) {
	svc := newTestAccountService(t)

	_, err := svc.Get(context.Background(), []uint{7}, 9999)
	require.ErrorIs(t, err, ErrAccountNotFound)
}

func TestAccountPersistenceUpdateName(t *testing.T) {
	svc := newTestAccountService(t)

	row, err := svc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name:           "old-name",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)

	newName := "new-name"
	updated, err := svc.Update(context.Background(), []uint{7}, row.ID, cloudRequest.UpdateAccountRequest{Name: &newName})
	require.NoError(t, err)
	require.Equal(t, "new-name", updated.Name)
}

func TestAccountPersistenceSoftDelete(t *testing.T) {
	svc := newTestAccountService(t)
	db := global.OPTIMUS_DB

	row, err := svc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(context.Background(), []uint{7}, row.ID))

	_, err = svc.Get(context.Background(), []uint{7}, row.ID)
	require.ErrorIs(t, err, ErrAccountNotFound)

	var retained, deleted int64
	require.NoError(t, db.Unscoped().Table("cloud_accounts").Where("id = ?", row.ID).Count(&retained).Error)
	require.NoError(t, db.Unscoped().Table("cloud_accounts").Where("id = ? AND deleted_at IS NOT NULL", row.ID).Count(&deleted).Error)
	require.Equal(t, int64(1), retained)
	require.Equal(t, int64(1), deleted)
}

func TestAccountPersistenceGetRejectsAccountInOtherProject(t *testing.T) {
	svc := newTestAccountService(t)

	row, err := svc.Create(context.Background(), 7, cloudRequest.CreateAccountRequest{
		Name:           "prod",
		Provider:       "aliyun",
		AccountID:      "123",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)

	_, err = svc.Get(context.Background(), []uint{8}, row.ID)
	require.ErrorIs(t, err, ErrAccountNotFound)
}

func TestAccountPersistenceUpdateTagsRoundTrip(t *testing.T) {
	svc := newTestAccountService(t)

	row, err := svc.Create(context.Background(), 1, cloudRequest.CreateAccountRequest{
		Name:           "n",
		Provider:       "aliyun",
		AccountID:      "a",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)

	updated, err := svc.UpdateTags(context.Background(), []uint{1}, row.ID, cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "prod", "team": "sre"}})
	require.NoError(t, err)
	require.Equal(t, "prod", updated.Tags["environment"])
	require.Equal(t, "sre", updated.Tags["team"])

	got, err := svc.Get(context.Background(), []uint{1}, row.ID)
	require.NoError(t, err)
	require.Equal(t, "prod", got.Tags["environment"])
}

func TestAccountPersistenceListTagFilterEnvironmentProd(t *testing.T) {
	svc := newTestAccountService(t)

	r1, err := svc.Create(context.Background(), 1, cloudRequest.CreateAccountRequest{
		Name:           "p1",
		Provider:       "aliyun",
		AccountID:      "a1",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)
	_, err = svc.UpdateTags(context.Background(), []uint{1}, r1.ID, cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "prod"}})
	require.NoError(t, err)

	r2, err := svc.Create(context.Background(), 1, cloudRequest.CreateAccountRequest{
		Name:           "p2",
		Provider:       "aliyun",
		AccountID:      "a2",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)
	_, err = svc.UpdateTags(context.Background(), []uint{1}, r2.ID, cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "dev"}})
	require.NoError(t, err)

	rows, err := svc.List(context.Background(), []uint{1},
		[]TagFilter{{Key: "environment", Value: "prod", Exists: true}})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "p1", rows[0].Name)
}

func TestAccountPersistenceListTagFilterNoneKey(t *testing.T) {
	svc := newTestAccountService(t)

	r1, err := svc.Create(context.Background(), 1, cloudRequest.CreateAccountRequest{
		Name:           "tagged",
		Provider:       "aliyun",
		AccountID:      "a1",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)
	_, err = svc.UpdateTags(context.Background(), []uint{1}, r1.ID, cloudRequest.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "prod"}})
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), 1, cloudRequest.CreateAccountRequest{
		Name:           "untagged",
		Provider:       "aliyun",
		AccountID:      "a2",
		CredentialType: "ak_sk",
	})
	require.NoError(t, err)

	rows, err := svc.List(context.Background(), []uint{1},
		[]TagFilter{{Key: "environment", Exists: false}})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "untagged", rows[0].Name)
}
