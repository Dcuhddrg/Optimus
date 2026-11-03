package project

import (
	"context"
	"testing"

	projectRequest "optimus/server/model/project/request"

	"github.com/stretchr/testify/require"
)

func TestCreate_PersistsOwnerMembership(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	svc := &ProjectService{}

	p, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "Prod", Code: "prod"})
	require.NoError(t, err)
	require.NotZero(t, p.ID)
	require.Equal(t, uint(1), p.OwnerID)

	m, err := svc.findMember(context.Background(), 1, p.ID)
	require.NoError(t, err)
	require.Equal(t, RoleOwner, m.ProjectRole)
}

func TestCreate_DuplicateCodeReturnsCodeTaken(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	svc := &ProjectService{}

	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "prod"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P2", Code: "prod"})
	require.ErrorIs(t, err, ErrProjectCodeTaken)
}

func TestListMembers_JoinsSystemUsers(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}

	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "Prod", Code: "prod"})
	require.NoError(t, err)
	require.NoError(t, svc.AddMember(context.Background(), 1, "prod", 2, RoleMember))

	members, err := svc.ListMembers(context.Background(), 1, "prod")
	require.NoError(t, err)
	require.Len(t, members, 2)

	roles := map[string]string{}
	for _, m := range members {
		roles[m.Username] = m.ProjectRole
	}
	require.Equal(t, RoleOwner, roles["alice"])
	require.Equal(t, RoleMember, roles["bob"])
}

func TestListAccessibleProjectIDs_ReturnsOnlyMemberships(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}

	p1, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 2, projectRequest.CreateProjectRequest{Name: "P2", Code: "p2"})
	require.NoError(t, err)

	ids, err := svc.ListAccessibleProjectIDs(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []uint{p1.ID}, ids)
}

func TestAddMember_ChecksUserExists(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	svc := &ProjectService{}

	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P", Code: "p1"})
	require.NoError(t, err)
	require.NoError(t, svc.AddMember(context.Background(), 1, "p1", 1, RoleOwner))
	require.ErrorIs(t, svc.AddMember(context.Background(), 1, "p1", 999, RoleMember), ErrUserNotFound)
}

func TestCountOwners(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}

	p, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P", Code: "p1"})
	require.NoError(t, err)

	n, err := svc.countOwners(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	require.NoError(t, svc.AddMember(context.Background(), 1, "p1", 2, RoleOwner))
	n, err = svc.countOwners(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
}

func TestUpdate_UnchangedFieldsStillSucceeds(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P", Code: "p1", Description: "x"})
	require.NoError(t, err)

	// Update with same description — must NOT return ErrProjectNotFound under MySQL.
	description := "x"
	updated, err := svc.Update(context.Background(), 1, "p1", projectRequest.UpdateProjectRequest{Description: &description})
	require.NoError(t, err)
	require.Equal(t, "x", updated.Description)
}

func TestUpdateMemberRole_SameRoleStillSucceeds(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	svc := &ProjectService{}
	_, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P", Code: "p1"})
	require.NoError(t, err)

	// Owner→Owner: same role, must NOT return ErrMemberNotFound.
	require.NoError(t, svc.UpdateMemberRole(context.Background(), 1, "p1", 1, RoleOwner))
}

func TestDelete_RemovesOrphanUserProjects(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	p, err := svc.Create(context.Background(), 1, projectRequest.CreateProjectRequest{Name: "P", Code: "p1"})
	require.NoError(t, err)
	require.NoError(t, svc.AddMember(context.Background(), 1, "p1", 2, RoleMember))

	require.NoError(t, svc.Delete(context.Background(), 1, "p1"))

	// All user_projects rows for this project should be gone.
	_, err = svc.ListMembers(context.Background(), 1, "p1")
	require.ErrorIs(t, err, ErrProjectNotFound)

	var retained, deleted, memberships int64
	require.NoError(t, db.Unscoped().Table("projects").Where("id = ?", p.ID).Count(&retained).Error)
	require.NoError(t, db.Unscoped().Table("projects").Where("id = ? AND deleted_at IS NOT NULL", p.ID).Count(&deleted).Error)
	require.NoError(t, db.Unscoped().Table("user_projects").Where("project_id = ?", p.ID).Count(&memberships).Error)
	require.Equal(t, int64(1), retained)
	require.Equal(t, int64(1), deleted)
	require.Zero(t, memberships)
}
