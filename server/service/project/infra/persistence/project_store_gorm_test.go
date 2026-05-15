package persistence

import (
	"context"
	"testing"

	projectModel "optimus/server/model/project"
	"optimus/server/service/project/usecase"

	"github.com/stretchr/testify/require"
)

func TestCreate_PersistsOwnerMembership(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)

	p := &projectModel.Project{Name: "Prod", Code: "prod", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))
	require.NotZero(t, p.ID)
	require.Equal(t, uint(1), p.OwnerID)

	m, err := store.FindMember(context.Background(), 1, p.ID)
	require.NoError(t, err)
	require.Equal(t, usecase.RoleOwner, m.ProjectRole)
}

func TestCreate_DuplicateCodeReturnsCodeTaken(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)

	require.NoError(t, store.Create(context.Background(), &projectModel.Project{Name: "P1", Code: "prod", Status: "active"}, 1))
	err := store.Create(context.Background(), &projectModel.Project{Name: "P2", Code: "prod", Status: "active"}, 1)
	require.ErrorIs(t, err, usecase.ErrProjectCodeTaken)
}

func TestListMembers_JoinsSystemUsers(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)

	p := &projectModel.Project{Name: "Prod", Code: "prod", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))
	require.NoError(t, store.AddOrReplaceMember(context.Background(), projectModel.UserProject{
		UserID: 2, ProjectID: p.ID, ProjectRole: usecase.RoleMember,
	}))

	members, err := store.ListMembers(context.Background(), p.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)

	roles := map[string]string{}
	for _, m := range members {
		roles[m.Username] = m.ProjectRole
	}
	require.Equal(t, usecase.RoleOwner, roles["alice"])
	require.Equal(t, usecase.RoleMember, roles["bob"])
}

func TestListAccessibleProjectIDs_ReturnsOnlyMemberships(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)

	p1 := &projectModel.Project{Name: "P1", Code: "p1", Status: "active"}
	p2 := &projectModel.Project{Name: "P2", Code: "p2", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p1, 1))
	require.NoError(t, store.Create(context.Background(), p2, 2))

	ids, err := store.ListAccessibleProjectIDs(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []uint{p1.ID}, ids)
}

func TestUserExists(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)

	ok, err := store.UserExists(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = store.UserExists(context.Background(), 999)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestCountOwners(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)

	p := &projectModel.Project{Name: "P", Code: "p", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))

	n, err := store.CountOwners(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	require.NoError(t, store.AddOrReplaceMember(context.Background(), projectModel.UserProject{
		UserID: 2, ProjectID: p.ID, ProjectRole: usecase.RoleOwner,
	}))
	n, err = store.CountOwners(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
}

func TestUpdate_UnchangedFieldsStillSucceeds(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)
	p := &projectModel.Project{Name: "P", Code: "p", Status: "active", Description: "x"}
	require.NoError(t, store.Create(context.Background(), p, 1))

	// Update with same description — must NOT return ErrProjectNotFound under MySQL.
	updated, err := store.Update(context.Background(), "p", map[string]any{"description": "x"})
	require.NoError(t, err)
	require.Equal(t, "x", updated.Description)
}

func TestUpdateMemberRole_SameRoleStillSucceeds(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)
	p := &projectModel.Project{Name: "P", Code: "p", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))

	// Owner→Owner: same role, must NOT return ErrMemberNotFound.
	require.NoError(t, store.UpdateMemberRole(context.Background(), 1, p.ID, usecase.RoleOwner))
}

func TestDelete_RemovesOrphanUserProjects(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)
	p := &projectModel.Project{Name: "P", Code: "p", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))
	require.NoError(t, store.AddOrReplaceMember(context.Background(), projectModel.UserProject{
		UserID: 2, ProjectID: p.ID, ProjectRole: usecase.RoleMember,
	}))

	require.NoError(t, store.Delete(context.Background(), "p"))

	// All user_projects rows for this project should be gone.
	members, err := store.ListMembers(context.Background(), p.ID)
	require.NoError(t, err)
	require.Empty(t, members)
}
