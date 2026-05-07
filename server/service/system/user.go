package system

import (
	"context"
	"errors"

	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrLastAdminCapableUser = errors.New("must keep at least one enabled admin-capable user")
)

type UserStore interface {
	ListUsers(ctx context.Context) ([]systemModel.SystemUser, error)
	FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error)
	RolesForUser(ctx context.Context, userID uint) ([]SystemRoleSummary, error)
	UpdateUser(ctx context.Context, user systemModel.SystemUser) (systemModel.SystemUser, error)
	WouldHaveAdminCapableUserAfterStatus(ctx context.Context, userID uint, status string) (bool, error)
	WouldHaveAdminCapableUserAfterRoles(ctx context.Context, userID uint, roleIDs []uint) (bool, error)
	ReplaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error
	WriteAudit(ctx context.Context, event AuditEvent) error
}

type UserService struct {
	store UserStore
}

type UserListResult struct {
	Items []UserListItem `json:"items"`
	Total int            `json:"total"`
}

type UserListItem struct {
	SystemUserSummary
	Roles []string `json:"roles"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

func NewUserService(store UserStore) *UserService {
	return &UserService{store: store}
}

func (s *UserService) ListUsers(ctx context.Context) (UserListResult, error) {
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return UserListResult{}, err
	}

	items := make([]UserListItem, 0, len(users))
	for _, user := range users {
		roles, err := s.store.RolesForUser(ctx, user.ID)
		if err != nil {
			return UserListResult{}, err
		}
		_, roleCodes := summarizeRoles(roles)
		items = append(items, UserListItem{
			SystemUserSummary: summarizeUser(user),
			Roles:             roleCodes,
		})
	}

	return UserListResult{Items: items, Total: len(items)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, actorID uint, userID uint, req UpdateUserRequest, meta AuditMeta) (SystemUserSummary, error) {
	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return SystemUserSummary{}, err
	}
	user.DisplayName = req.DisplayName
	user.Email = req.Email

	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		return SystemUserSummary{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, Action: "user.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, actorID uint, userID uint, status string, meta AuditMeta) (SystemUserSummary, error) {
	ok, err := s.store.WouldHaveAdminCapableUserAfterStatus(ctx, userID, status)
	if err != nil {
		return SystemUserSummary{}, err
	}
	if !ok {
		return SystemUserSummary{}, ErrLastAdminCapableUser
	}

	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return SystemUserSummary{}, err
	}
	user.Status = status
	updated, err := s.store.UpdateUser(ctx, user)
	if err != nil {
		return SystemUserSummary{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, Action: "user.status.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
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
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, Action: "user.roles.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return nil
}

type GormUserStore struct {
	db *gorm.DB
}

func NewGormUserStore(db *gorm.DB) *GormUserStore {
	return &GormUserStore{db: db}
}

func (s *GormUserStore) ListUsers(ctx context.Context) ([]systemModel.SystemUser, error) {
	var users []systemModel.SystemUser
	err := s.db.WithContext(ctx).Order("id ASC").Find(&users).Error
	return users, err
}

func (s *GormUserStore) FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemUser{}, ErrUserNotFound
	}
	return user, err
}

func (s *GormUserStore) RolesForUser(ctx context.Context, userID uint) ([]SystemRoleSummary, error) {
	return NewGormAuthStore(s.db).RolesForUser(ctx, userID)
}

func (s *GormUserStore) UpdateUser(ctx context.Context, user systemModel.SystemUser) (systemModel.SystemUser, error) {
	err := s.db.WithContext(ctx).Save(&user).Error
	return user, err
}

func (s *GormUserStore) WouldHaveAdminCapableUserAfterStatus(ctx context.Context, userID uint, status string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Table("sys_users AS users").
		Joins("JOIN sys_user_roles user_roles ON user_roles.user_id = users.id").
		Joins("JOIN sys_roles roles ON roles.id = user_roles.role_id").
		Where("roles.code = ? AND roles.status = ?", "admin", defaultRoleStatus).
		Where("(users.id <> ? AND users.status = ?) OR (users.id = ? AND ? = ?)", userID, defaultUserStatus, userID, status, defaultUserStatus).
		Count(&count).Error
	return count > 0, err
}

func (s *GormUserStore) WouldHaveAdminCapableUserAfterRoles(ctx context.Context, userID uint, roleIDs []uint) (bool, error) {
	var adminRole systemModel.SystemRole
	if err := s.db.WithContext(ctx).Where("code = ? AND status = ?", "admin", defaultRoleStatus).First(&adminRole).Error; err != nil {
		return false, err
	}

	hasAdminRole := false
	for _, roleID := range roleIDs {
		if roleID == adminRole.ID {
			hasAdminRole = true
			break
		}
	}

	var count int64
	err := s.db.WithContext(ctx).
		Table("sys_users AS users").
		Joins("JOIN sys_user_roles user_roles ON user_roles.user_id = users.id").
		Joins("JOIN sys_roles roles ON roles.id = user_roles.role_id").
		Where("users.id <> ? AND users.status = ? AND roles.code = ? AND roles.status = ?", userID, defaultUserStatus, "admin", defaultRoleStatus).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	var user systemModel.SystemUser
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.Status == defaultUserStatus && hasAdminRole, nil
}

func (s *GormUserStore) ReplaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&systemModel.SystemUserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&systemModel.SystemUserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GormUserStore) WriteAudit(ctx context.Context, event AuditEvent) error {
	return NewGormAuthStore(s.db).WriteAudit(ctx, event)
}
