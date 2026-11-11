package system

import (
	"errors"
	"strings"

	systemModel "optimus/server/model/system"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

const (
	DefaultAdminUsername = "admin"
	// BuiltinAdminUUID identifies the built-in local admin account, the
	// permanent fallback administrator (design 3.4). Identify it by UUID, not
	// by username.
	BuiltinAdminUUID = "00000000-0000-0000-0000-000000000001"
	// AdminRoleCode is the built-in role that always holds every system.*
	// permission.
	AdminRoleCode = "admin"
	// DefaultUserRoleCode is bound to new LDAP employees and cannot be deleted.
	DefaultUserRoleCode = "user"
	DefaultUserStatus   = "enabled"
)

func AutoMigrateSystem(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	return db.AutoMigrate(
		&systemModel.SystemUser{},
		&systemModel.SystemRole{},
		&systemModel.SystemUserRole{},
		&systemModel.SystemPermission{},
		&systemModel.SystemRolePermission{},
		&systemModel.SystemLDAPConfig{},
		&systemModel.SystemAuditLog{},
		&systemModel.SystemDataMigration{},
		&systemModel.JwtBlacklist{},
		&gormadapter.CasbinRule{},
	)
}

// RetiredSystemPermissions are catalog entries removed from the product;
// startup deletes them together with their grants.
func RetiredSystemPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "system.roles.status.update", Type: "api", Method: "PATCH", Path: "/system/roles/:id/status"},
	}
}

func DefaultSystemRoles() []systemModel.SystemRole {
	return []systemModel.SystemRole{
		{Code: "admin", Name: "Administrator", Description: "Full system administrator"},
		{Code: "user", Name: "User", Description: "Default LDAP user role"},
		// Approval roles carry no permissions; they only qualify holders for
		// an approval level (design 3.5).
		{Code: "security-approver", Name: "Security approver", Description: "Approves the security level", Approvable: true},
		{Code: "dept-head-approver", Name: "Department head approver", Description: "Approves the department head level", Approvable: true},
	}
}

// SystemPermissionCodes lists the system.* catalog codes the admin role must
// always hold.
func SystemPermissionCodes() []string {
	codes := []string{}
	for _, permission := range DefaultSystemPermissions() {
		if strings.HasPrefix(permission.Code, "system.") {
			codes = append(codes, permission.Code)
		}
	}
	return codes
}

func DefaultSystemPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "dashboard.view", Name: "Dashboard", Type: "menu", Path: "/", Sort: 10},
		{Code: "system.menu", Name: "System", Type: "menu", Path: "/system", Sort: 20},
		{Code: "system.users.menu", Name: "Users", Type: "menu", Path: "/system/users", ParentCode: "system.menu", Sort: 21},
		{Code: "system.roles.menu", Name: "Roles", Type: "menu", Path: "/system/roles", ParentCode: "system.menu", Sort: 22},
		{Code: "system.ldap.menu", Name: "LDAP", Type: "menu", Path: "/system/ldap", ParentCode: "system.menu", Sort: 23},
		{Code: "system.audit.menu", Name: "Audit Logs", Type: "menu", Path: "/system/audit-logs", ParentCode: "system.menu", Sort: 24},
		{Code: "system.auth.me", Name: "Current user", Type: "api", Method: "GET", Path: "/system/auth/me", Sort: 100},
		{Code: "system.auth.logout", Name: "Logout", Type: "api", Method: "POST", Path: "/system/auth/logout", Sort: 101},
		{Code: "system.users.view", Name: "View users", Type: "api", Method: "GET", Path: "/system/users", Sort: 200},
		{Code: "system.users.update", Name: "Update users", Type: "api", Method: "PUT", Path: "/system/users/:id", Sort: 201},
		{Code: "system.users.status.update", Name: "Update user status", Type: "api", Method: "PATCH", Path: "/system/users/:id/status", Sort: 202},
		{Code: "system.users.roles.update", Name: "Update user roles", Type: "api", Method: "PUT", Path: "/system/users/:id/roles", Sort: 203},
		{Code: "system.roles.view", Name: "View roles", Type: "api", Method: "GET", Path: "/system/roles", Sort: 300},
		{Code: "system.roles.create", Name: "Create roles", Type: "api", Method: "POST", Path: "/system/roles", Sort: 301},
		{Code: "system.roles.update", Name: "Update roles", Type: "api", Method: "PUT", Path: "/system/roles/:id", Sort: 302},
		{Code: "system.roles.delete", Name: "Delete roles", Type: "api", Method: "DELETE", Path: "/system/roles/:id", Sort: 303},
		{Code: "system.roles.permissions.update", Name: "Update role permissions", Type: "api", Method: "PUT", Path: "/system/roles/:id/permissions", Sort: 304},
		{Code: "system.permissions.view", Name: "View permissions", Type: "api", Method: "GET", Path: "/system/permissions/tree", Sort: 400},
		{Code: "system.ldap.view", Name: "View LDAP config", Type: "api", Method: "GET", Path: "/system/ldap/config", Sort: 500},
		{Code: "system.ldap.update", Name: "Update LDAP config", Type: "api", Method: "PUT", Path: "/system/ldap/config", Sort: 501},
		{Code: "system.ldap.test", Name: "Test LDAP config", Type: "api", Method: "POST", Path: "/system/ldap/test", Sort: 502},
		{Code: "system.audit.view", Name: "View audit logs", Type: "api", Method: "GET", Path: "/system/audit-logs", Sort: 600},
		{Code: "system.audit.export", Name: "Export audit logs", Type: "api", Method: "GET", Path: "/system/audit-logs/export", Sort: 601},
	}
}
