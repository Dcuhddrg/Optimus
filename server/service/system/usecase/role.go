package usecase

import (
	"context"
	"errors"

	systemModel "optimus/server/model/system"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrProtectedAdminRole = errors.New("admin role is protected")
)

type RoleStore interface {
	ListRoles(ctx context.Context) ([]systemModel.SystemRole, error)
	FindRoleByID(ctx context.Context, id uint) (systemModel.SystemRole, error)
	CreateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error)
	UpdateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error)
	PermissionsByCodes(ctx context.Context, codes []string) ([]systemModel.SystemPermission, error)
	ReplaceRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	WriteAudit(ctx context.Context, event AuditEvent) error
}

type RoleService struct {
	store RoleStore
}

type RoleRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleListResult struct {
	Items []systemModel.SystemRole `json:"items"`
	Total int                      `json:"total"`
}

func NewRoleService(store RoleStore) *RoleService {
	return &RoleService{store: store}
}

func (s *RoleService) ListRoles(ctx context.Context) ([]systemModel.SystemRole, error) {
	return s.store.ListRoles(ctx)
}

func (s *RoleService) CreateRole(ctx context.Context, actorID uint, req RoleRequest, meta AuditMeta) (systemModel.SystemRole, error) {
	role := systemModel.SystemRole{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      defaultRoleStatus,
	}
	created, err := s.store.CreateRole(ctx, role)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.create", ResourceType: "role", ResourceID: resourceID(created.ID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return created, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, actorID uint, roleID uint, req RoleRequest, meta AuditMeta) (systemModel.SystemRole, error) {
	role, err := s.store.FindRoleByID(ctx, roleID)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	if role.Code == "admin" && req.Code != "" && req.Code != "admin" {
		return systemModel.SystemRole{}, ErrProtectedAdminRole
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	role.Name = req.Name
	role.Description = req.Description
	updated, err := s.store.UpdateRole(ctx, role)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.update", ResourceType: "role", ResourceID: resourceID(roleID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return updated, nil
}

func (s *RoleService) UpdateRoleStatus(ctx context.Context, actorID uint, roleID uint, status string, meta AuditMeta) (systemModel.SystemRole, error) {
	role, err := s.store.FindRoleByID(ctx, roleID)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	if role.Code == "admin" && status != defaultRoleStatus {
		return systemModel.SystemRole{}, ErrProtectedAdminRole
	}
	role.Status = status
	updated, err := s.store.UpdateRole(ctx, role)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.status.update", ResourceType: "role", ResourceID: resourceID(roleID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return updated, nil
}

func (s *RoleService) UpdateRolePermissions(ctx context.Context, actorID uint, roleID uint, permissionCodes []string, meta AuditMeta) error {
	role, err := s.store.FindRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role.Code == "admin" && !containsAllCoreAdminPermissions(permissionCodes) {
		return ErrProtectedAdminRole
	}
	permissions, err := s.store.PermissionsByCodes(ctx, permissionCodes)
	if err != nil {
		return err
	}
	permissionIDs := make([]uint, 0, len(permissions))
	for _, permission := range permissions {
		permissionIDs = append(permissionIDs, permission.ID)
	}
	if err := s.store.ReplaceRolePermissions(ctx, roleID, permissionIDs); err != nil {
		return err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.permissions.update", ResourceType: "role", ResourceID: resourceID(roleID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return nil
}

func containsAllCoreAdminPermissions(codes []string) bool {
	set := make(map[string]bool, len(codes))
	for _, code := range codes {
		set[code] = true
	}
	for _, required := range []string{
		"system.users.view",
		"system.roles.permissions.update",
		"system.ldap.update",
		"system.audit.view",
		"system.audit.export",
	} {
		if !set[required] {
			return false
		}
	}
	return true
}
