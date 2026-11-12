package project

import "errors"

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
	// ErrExternalCannotOwn: external accounts cannot be project owners
	// (user management design 3.6).
	ErrExternalCannotOwn = errors.New("external accounts cannot be project owners")
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
