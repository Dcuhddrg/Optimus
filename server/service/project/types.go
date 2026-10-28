package project

import (
	"context"
	"errors"

	projectModel "optimus/server/model/project"
)

var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrProjectCodeTaken      = errors.New("project code already taken")
	ErrInvalidProjectCode    = errors.New("project code must match ^[a-z][a-z0-9-]{1,62}$")
	ErrInvalidProjectName    = errors.New("project name must be 1-128 chars")
	ErrProjectHasAccounts    = errors.New("project has bound cloud accounts; unbind first")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidProjectRole    = errors.New("project role must be owner/member/viewer")
	ErrMemberNotFound        = errors.New("member not found")
	ErrCannotRemoveLastOwner = errors.New("cannot remove the last project owner")
	ErrUserNotFound          = errors.New("target user does not exist")
)

const (
	RoleOwner  = "owner"
	RoleMember = "member"
	RoleViewer = "viewer"

	statusActive = "active"
)

func IsValidProjectRole(role string) bool {
	return role == RoleOwner || role == RoleMember || role == RoleViewer
}

type ProjectStore interface {
	Create(ctx context.Context, p *projectModel.Project, ownerID uint) error
	FindByCode(ctx context.Context, code string) (projectModel.Project, error)
	List(ctx context.Context, accessibleIDs []uint) ([]projectModel.Project, error)
	Update(ctx context.Context, code string, fields map[string]any) (projectModel.Project, error)
	Delete(ctx context.Context, code string) error
	CountCloudAccounts(ctx context.Context, projectID uint) (int64, error)

	AddOrReplaceMember(ctx context.Context, m projectModel.UserProject) error
	UpdateMemberRole(ctx context.Context, userID, projectID uint, role string) error
	RemoveMember(ctx context.Context, userID, projectID uint) error
	ListMembers(ctx context.Context, projectID uint) ([]projectModel.ProjectMember, error)
	FindMember(ctx context.Context, userID, projectID uint) (projectModel.UserProject, error)
	ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error)
	CountOwners(ctx context.Context, projectID uint) (int64, error)
	UserExists(ctx context.Context, userID uint) (bool, error)
}
