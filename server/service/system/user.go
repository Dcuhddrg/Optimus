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
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, actorID uint, userID uint, status string, meta AuditMeta) (systemModel.SystemUserSummary, error) {
	ok, err := s.store.WouldHaveAdminCapableUserAfterStatus(ctx, userID, status)
	if err != nil {
		return systemModel.SystemUserSummary{}, err
	}
	if !ok {
		return systemModel.SystemUserSummary{}, ErrLastAdminCapableUser
	}

	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return systemModel.SystemUserSummary{}, err
	}
	user.Status = status
	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		return systemModel.SystemUserSummary{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.status.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint, meta AuditMeta) error {
	ok, err := s.store.WouldHaveAdminCapableUserAfterRoles(ctx, userID, roleIDs)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLastAdminCapableUser
	}
	if err := s.store.ReplaceUserRoles(ctx, userID, roleIDs); err != nil {
		return err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.roles.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return nil
}
