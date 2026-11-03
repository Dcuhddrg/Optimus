package project

import (
	"context"
	"testing"

	cloudModel "optimus/server/model/cloud"
	projectRequest "optimus/server/model/project/request"

	"github.com/stretchr/testify/require"
)

func TestCreate_AssignsActorAsOwner(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}

	view, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{
		Name: "Prod", Code: "prod", Description: "Production",
	})
	require.NoError(t, err)
	require.Equal(t, "prod", view.Code)
	require.Equal(t, uint(1), view.OwnerID)

	members, err := svc.ListMembers(context.Background(), 1, "prod")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, RoleOwner, members[0].ProjectRole)
}

func TestCreate_RejectsInvalidCode(t *testing.T) {
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "X", Code: "Prod"})
	require.ErrorIs(t, err, ErrInvalidProjectCode)

	_, err = svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "X", Code: "1bad"})
	require.ErrorIs(t, err, ErrInvalidProjectCode)
}

func TestCreate_RejectsDuplicateCode(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}

	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "prod"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P2", Code: "prod"})
	require.ErrorIs(t, err, ErrProjectCodeTaken)
}

func TestList_OnlyReturnsAccessible(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 2, projectRequest.CreateProjectRequest{Name: "P2", Code: "p2"})
	require.NoError(t, err)

	views, err := svc.List(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, "p1", views[0].Code)
}

func TestUpdate_OnlyOwnerCanUpdate(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	name := "P1-Renamed"
	_, err = svc.Update(context.Background(), 2, "p1", projectRequest.UpdateProjectRequest{Name: &name})
	require.ErrorIs(t, err, ErrForbidden)

	view, err := svc.Update(context.Background(), 1, "p1", projectRequest.UpdateProjectRequest{Name: &name})
	require.NoError(t, err)
	require.Equal(t, "P1-Renamed", view.Name)
}

func TestDelete_RejectsIfCloudAccountsBound(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	view, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	require.NoError(t, db.Create(&cloudModel.CloudAccount{
		ProjectID: view.ID, Name: "bound", Provider: "aliyun", AccountID: "account",
		CredentialType: "ak_sk", TagsJSON: "{}",
	}).Error)

	err = svc.Delete(context.Background(), 1, "p1")
	require.ErrorIs(t, err, ErrProjectHasAccounts)
}

func TestDelete_OnlyOwnerCanDelete(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.Delete(context.Background(), 2, "p1")
	require.ErrorIs(t, err, ErrForbidden)

	err = svc.Delete(context.Background(), 1, "p1")
	require.NoError(t, err)
}

func TestAddMember_OnlyOwnerCanAdd(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	seedUser(t, db, 3, "carol")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 2, "p1", 3, RoleMember)
	require.ErrorIs(t, err, ErrForbidden)

	err = svc.AddMember(context.Background(), 1, "p1", 2, RoleMember)
	require.NoError(t, err)
}

func TestAddMember_RejectsInvalidRole(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 1, "p1", 2, "admin")
	require.ErrorIs(t, err, ErrInvalidProjectRole)
}

func TestAddMember_RejectsUnknownUser(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 1, "p1", 999, RoleMember)
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestRemoveMember_PreventsRemovingLastOwner(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.RemoveMember(context.Background(), 1, "p1", 1)
	require.ErrorIs(t, err, ErrCannotRemoveLastOwner)
}

func TestUpdateMemberRole_PreventsDemotingLastOwner(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.UpdateMemberRole(context.Background(), 1, "p1", 1, RoleMember)
	require.ErrorIs(t, err, ErrCannotRemoveLastOwner)
}
