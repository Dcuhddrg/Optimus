package system

import (
	"context"
	"errors"

	systemModel "optimus/server/model/system"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrLastAdminCapableUser = errors.New("must keep at least one enabled admin-capable user")
)

type UserStore interface {
	ListUsers(ctx context.Context) ([]systemModel.SystemUser, error)
	FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error)
	RolesForUser(ctx context.Context, userID uint) ([]systemModel.SystemRoleSummary, error)
	UpdateUser(ctx context.Context, user systemModel.SystemUser) (systemModel.SystemUser, error)
	WouldHaveAdminCapableUserAfterStatus(ctx context.Context, userID uint, status string) (bool, error)
	WouldHaveAdminCapableUserAfterRoles(ctx context.Context, userID uint, roleIDs []uint) (bool, error)
	ReplaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error
	WriteAudit(ctx context.Context, event AuditEvent) error
}

type UserService struct {
	store UserStore
}

func NewUserService(store UserStore) *UserService {
	return &UserService{store: store}
}

func (s *UserService) ListUsers(ctx context.Context) (systemModel.UserListResult, error) {
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return systemModel.UserListResult{}, err
	}

	items := make([]systemModel.UserListItem, 0, len(users))
	for _, user := range users {
		roles, err := s.store.RolesForUser(ctx, user.ID)
		if err != nil {
			return systemModel.UserListResult{}, err
		}
		_, roleCodes := summarizeRoles(roles)
		items = append(items, systemModel.UserListItem{
			SystemUserSummary: summarizeUser(user),
			Roles:             roleCodes,
		})
	}

	return systemModel.UserListResult{Items: items, Total: len(items)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, actorID uint, userID uint, req systemModel.UpdateUserRequest, meta AuditMeta) (systemModel.SystemUserSummary, error) {
	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return systemModel.SystemUserSummary{}, err
	}
	user.DisplayName = req.DisplayName
	user.Email = req.Email

	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		return systemModel.SystemUserSummary{}, err
	}
	recordAudit(ctx, s.store, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, actorID uint, userID uint, status string, meta AuditMeta) (systemModel.SystemUserSummary, error) {
	event := failedAudit(actorID, meta, "user.status.update", "user", userID)
	defer func() { recordAudit(ctx, s.store, event) }()
	ok, err := s.store.WouldHaveAdminCapableUserAfterStatus(ctx, userID, status)
	if err != nil {
		event.Detail = auditDetails("admin_check_failed", nil)
		return systemModel.SystemUserSummary{}, err
	}
	if !ok {
		event.Detail = auditDetails("last_admin_protected", nil)
		return systemModel.SystemUserSummary{}, ErrLastAdminCapableUser
	}

	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return systemModel.SystemUserSummary{}, err
	}
	before := user.Status
	user.Status = status
	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemModel.SystemUserSummary{}, err
	}
	event.Result = "success"
	event.Detail = auditDetails("updated", map[string]auditChange{"status": {Before: before, After: updated.Status}})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint, meta AuditMeta) error {
	event := failedAudit(actorID, meta, "user.roles.update", "user", userID)
	defer func() { recordAudit(ctx, s.store, event) }()
	ok, err := s.store.WouldHaveAdminCapableUserAfterRoles(ctx, userID, roleIDs)
	if err != nil {
		event.Detail = auditDetails("admin_check_failed", nil)
		return err
	}
	if !ok {
		event.Detail = auditDetails("last_admin_protected", nil)
		return ErrLastAdminCapableUser
	}
	before, beforeErr := s.store.RolesForUser(ctx, userID)
	if err := s.store.ReplaceUserRoles(ctx, userID, roleIDs); err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return err
	}
	after, afterErr := s.store.RolesForUser(ctx, userID)
	event.Result = "success"
	if beforeErr != nil || afterErr != nil {
		event.Detail = auditDetails("snapshot_unavailable", nil)
	} else {
		beforeIDs, afterIDs := make([]uint, 0, len(before)), make([]uint, 0, len(after))
		for _, role := range before {
			beforeIDs = append(beforeIDs, role.ID)
		}
		for _, role := range after {
			afterIDs = append(afterIDs, role.ID)
		}
		event.Detail = auditDetails("updated", map[string]auditChange{"roleIds": auditSetChange(beforeIDs, afterIDs)})
	}
	return nil
}
