package system

import (
	"errors"

	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

const (
	defaultAdminUsername = "admin"
	defaultUserStatus    = "enabled"
	defaultRoleStatus    = "enabled"
	defaultAdminPassword = "admin"
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
	)
}

func SeedSystemData(db *gorm.DB, initialAdminPassword string) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if initialAdminPassword == "" {
		initialAdminPassword = defaultAdminPassword
	}

	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedPermissions(db); err != nil {
		return err
	}
	if err := seedAdminUser(db, initialAdminPassword); err != nil {
		return err
	}
	if err := seedAdminUserRole(db); err != nil {
		return err
	}
	return seedAdminRolePermissions(db)
}

func DefaultSystemRoles() []systemModel.SystemRole {
	return []systemModel.SystemRole{
		{Code: "admin", Name: "Administrator", Description: "Full system administrator", Status: defaultRoleStatus},
		{Code: "user", Name: "User", Description: "Default LDAP user role", Status: defaultRoleStatus},
	}
}

func DefaultSystemPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "dashboard.view", Name: "Dashboard", Type: "menu", Path: "/", Sort: 10},
		{Code: "system.menu", Name: "System", Type: "menu", Path: "/system", Sort: 20},
		{Code: "system.users.menu", Name: "Users", Type: "menu", Path: "/system/users", ParentCode: "system.menu", Sort: 21},
		{Code: "system.roles.menu", Name: "Roles", Type: "menu", Path: "/system/roles", ParentCode: "system.menu", Sort: 22},
		{Code: "system.ldap.menu", Name: "LDAP", Type: "menu", Path: "/system/ldap", ParentCode: "system.menu", Sort: 23},
		{Code: "system.audit.menu", Name: "Audit Logs", Type: "menu", Path: "/system/audit-logs", ParentCode: "system.menu", Sort: 24},
		{Code: "system.auth.me", Name: "Current user", Type: "api", Method: "GET", Path: "/api/v1/system/auth/me", Sort: 100},
		{Code: "system.auth.logout", Name: "Logout", Type: "api", Method: "POST", Path: "/api/v1/system/auth/logout", Sort: 101},
		{Code: "system.users.view", Name: "View users", Type: "api", Method: "GET", Path: "/api/v1/system/users", Sort: 200},
		{Code: "system.users.update", Name: "Update users", Type: "api", Method: "PUT", Path: "/api/v1/system/users/:id", Sort: 201},
		{Code: "system.users.status.update", Name: "Update user status", Type: "api", Method: "PATCH", Path: "/api/v1/system/users/:id/status", Sort: 202},
		{Code: "system.users.roles.update", Name: "Update user roles", Type: "api", Method: "PUT", Path: "/api/v1/system/users/:id/roles", Sort: 203},
		{Code: "system.roles.view", Name: "View roles", Type: "api", Method: "GET", Path: "/api/v1/system/roles", Sort: 300},
		{Code: "system.roles.create", Name: "Create roles", Type: "api", Method: "POST", Path: "/api/v1/system/roles", Sort: 301},
		{Code: "system.roles.update", Name: "Update roles", Type: "api", Method: "PUT", Path: "/api/v1/system/roles/:id", Sort: 302},
		{Code: "system.roles.status.update", Name: "Update role status", Type: "api", Method: "PATCH", Path: "/api/v1/system/roles/:id/status", Sort: 303},
		{Code: "system.roles.permissions.update", Name: "Update role permissions", Type: "api", Method: "PUT", Path: "/api/v1/system/roles/:id/permissions", Sort: 304},
		{Code: "system.permissions.view", Name: "View permissions", Type: "api", Method: "GET", Path: "/api/v1/system/permissions/tree", Sort: 400},
		{Code: "system.ldap.view", Name: "View LDAP config", Type: "api", Method: "GET", Path: "/api/v1/system/ldap/config", Sort: 500},
		{Code: "system.ldap.update", Name: "Update LDAP config", Type: "api", Method: "PUT", Path: "/api/v1/system/ldap/config", Sort: 501},
		{Code: "system.ldap.test", Name: "Test LDAP config", Type: "api", Method: "POST", Path: "/api/v1/system/ldap/test", Sort: 502},
		{Code: "system.audit.view", Name: "View audit logs", Type: "api", Method: "GET", Path: "/api/v1/system/audit-logs", Sort: 600},
		{Code: "system.audit.export", Name: "Export audit logs", Type: "api", Method: "GET", Path: "/api/v1/system/audit-logs/export", Sort: 601},
	}
}

func seedRoles(db *gorm.DB) error {
	for _, role := range DefaultSystemRoles() {
		if err := db.Where("code = ?", role.Code).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPermissions(db *gorm.DB) error {
	for _, permission := range DefaultSystemPermissions() {
		if err := db.Where("code = ?", permission.Code).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedAdminUser(db *gorm.DB, initialAdminPassword string) error {
	var existing systemModel.SystemUser
	err := db.Where("username = ?", defaultAdminUsername).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := HashPassword(initialAdminPassword)
	if err != nil {
		return err
	}

	admin := systemModel.SystemUser{
		UUID:         "00000000-0000-0000-0000-000000000001",
		Username:     defaultAdminUsername,
		DisplayName:  "Administrator",
		Source:       "local",
		Status:       defaultUserStatus,
		PasswordHash: hash,
	}
	return db.Create(&admin).Error
}

func seedAdminUserRole(db *gorm.DB) error {
	var admin systemModel.SystemUser
	if err := db.Where("username = ?", defaultAdminUsername).First(&admin).Error; err != nil {
		return err
	}

	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	userRole := systemModel.SystemUserRole{UserID: admin.ID, RoleID: adminRole.ID}
	return db.Where("user_id = ? AND role_id = ?", admin.ID, adminRole.ID).FirstOrCreate(&userRole).Error
}

func seedAdminRolePermissions(db *gorm.DB) error {
	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	var permissions []systemModel.SystemPermission
	if err := db.Find(&permissions).Error; err != nil {
		return err
	}

	for _, permission := range permissions {
		rolePermission := systemModel.SystemRolePermission{
			RoleID:       adminRole.ID,
			PermissionID: permission.ID,
		}
		if err := db.Where("role_id = ? AND permission_id = ?", adminRole.ID, permission.ID).
			FirstOrCreate(&rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}
