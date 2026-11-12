package project

import (
	"context"
	"testing"
	"time"

	projectRequest "optimus/server/model/project/request"
	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// External accounts cannot become project owners (user management design
// 3.6): not by creating a project, being added as owner or being promoted.
func TestExternalAccountsCannotOwnProjects(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "eve")
	require.NoError(t, db.Model(&systemModel.SystemUser{}).Where("id = ?", 2).Update("person_type", "external").Error)
	svc := &ProjectService{}
	ctx := context.Background()

	_, err := svc.Create(ctx, 2, projectRequest.CreateProjectRequest{Name: "Eve", Code: "eve-proj"})
	require.ErrorIs(t, err, ErrExternalCannotOwn)

	_, err = svc.Create(ctx, 1, projectRequest.CreateProjectRequest{Name: "Prod", Code: "prod"})
	require.NoError(t, err)
	require.ErrorIs(t, svc.AddMember(ctx, 1, "prod", 2, RoleOwner), ErrExternalCannotOwn)
	require.NoError(t, svc.AddMember(ctx, 1, "prod", 2, RoleMember), "members are fine")
	require.ErrorIs(t, svc.UpdateMemberRole(ctx, 1, "prod", 2, RoleOwner), ErrExternalCannotOwn)
	m, err := svc.findMember(ctx, 2, mustProjectID(t, svc, "prod"))
	require.NoError(t, err)
	require.Equal(t, RoleMember, m.ProjectRole)
}

func mustProjectID(t *testing.T, svc *ProjectService, code string) uint {
	t.Helper()
	p, err := svc.findByCode(context.Background(), code)
	require.NoError(t, err)
	return p.ID
}

// The person type check locks the user row in the transaction that writes
// the ownership, so a concurrent type change waits for it instead of slipping
// in between the check and the write.
func TestOwnerAssignmentHoldsTheUserRowAgainstTypeChange(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	svc := &ProjectService{}
	ctx := context.Background()
	_, err := svc.Create(ctx, 1, projectRequest.CreateProjectRequest{Name: "Prod", Code: "prod"})
	require.NoError(t, err)

	typeChanged := make(chan error, 1)
	var finishedWhileLocked bool
	ownerLockedHook = func(userID uint) {
		if userID != 2 {
			return
		}
		go func() {
			typeChanged <- db.Session(&gorm.Session{NewDB: true}).Exec("UPDATE sys_users SET person_type = 'external' WHERE id = 2").Error
		}()
		select {
		case <-typeChanged:
			finishedWhileLocked = true
		case <-time.After(500 * time.Millisecond):
		}
	}
	t.Cleanup(func() { ownerLockedHook = nil })

	require.NoError(t, svc.AddMember(ctx, 1, "prod", 2, RoleOwner))
	require.False(t, finishedWhileLocked, "the type change waited for the owner write")
	ownerLockedHook = nil
	require.NoError(t, <-typeChanged)

	// The type change landed after the ownership: the allowed "owner turned
	// external" state, which the sync lists for hand-over. A new assignment
	// is refused.
	m, err := svc.findMember(ctx, 2, mustProjectID(t, svc, "prod"))
	require.NoError(t, err)
	require.Equal(t, RoleOwner, m.ProjectRole)
	_, err = svc.Create(ctx, 2, projectRequest.CreateProjectRequest{Name: "Bob", Code: "bob-proj"})
	require.ErrorIs(t, err, ErrExternalCannotOwn)
}
