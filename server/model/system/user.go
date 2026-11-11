package system

import (
	"optimus/server/global"
	"time"
)

type SystemUser struct {
	global.OPTIMUS_MODEL
	UUID        string `gorm:"size:64;uniqueIndex;not null" json:"uuid"`
	Username    string `gorm:"size:128;uniqueIndex;not null" json:"username"`
	DisplayName string `gorm:"size:128" json:"displayName"`
	Email       string `gorm:"size:255" json:"email"`
	// Department is the LDAP value as is; display and filtering only, never
	// used for authorization.
	Department string `gorm:"size:128" json:"department"`
	// ExternalID is the immutable directory identity (id_attr) an LDAP account
	// is matched by; empty for local accounts and for legacy LDAP accounts
	// awaiting admin-confirmed binding.
	ExternalID   *string    `gorm:"size:255;uniqueIndex" json:"externalId"`
	Source       string     `gorm:"size:32;not null;index" json:"source"`
	Status       string     `gorm:"size:32;not null;index" json:"status"`
	PasswordHash string     `gorm:"size:255" json:"-"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
}

func (SystemUser) TableName() string {
	return "sys_users"
}

type SystemRole struct {
	global.OPTIMUS_MODEL
	Code        string `gorm:"size:128;uniqueIndex;not null" json:"code"`
	Name        string `gorm:"size:128;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	// Approvable marks a role usable as an approval level (design 3.5). As in
	// GVA, roles have no enabled/disabled status.
	Approvable bool `gorm:"not null;default:false" json:"approvable"`
}

func (SystemRole) TableName() string {
	return "sys_roles"
}

type SystemUserRole struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_sys_user_role" json:"userId"`
	RoleID    uint      `gorm:"not null;uniqueIndex:idx_sys_user_role" json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SystemUserRole) TableName() string {
	return "sys_user_roles"
}

type SystemPermission struct {
	global.OPTIMUS_MODEL
	Code       string `gorm:"size:160;uniqueIndex;not null" json:"code"`
	Name       string `gorm:"size:128;not null" json:"name"`
	Type       string `gorm:"size:32;not null;index" json:"type"`
	Method     string `gorm:"size:16;index" json:"method"`
	Path       string `gorm:"size:255;index" json:"path"`
	ParentCode string `gorm:"size:160;index" json:"parentCode"`
	Sort       int    `gorm:"not null;default:0" json:"sort"`
}

func (SystemPermission) TableName() string {
	return "sys_permissions"
}

type SystemRolePermission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RoleID       uint      `gorm:"not null;uniqueIndex:idx_sys_role_permission" json:"roleId"`
	PermissionID uint      `gorm:"not null;uniqueIndex:idx_sys_role_permission" json:"permissionId"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (SystemRolePermission) TableName() string {
	return "sys_role_permissions"
}

type SystemLDAPConfig struct {
	global.OPTIMUS_MODEL
	Enabled         bool   `gorm:"not null;default:false" json:"enabled"`
	URL             string `gorm:"size:255" json:"url"`
	BaseDN          string `gorm:"size:255" json:"baseDn"`
	BindDN          string `gorm:"size:255" json:"bindDn"`
	BindPassword    string `gorm:"size:255" json:"-"`
	UserFilter      string `gorm:"size:255" json:"userFilter"`
	UsernameAttr    string `gorm:"size:64" json:"usernameAttr"`
	DisplayNameAttr string `gorm:"size:64" json:"displayNameAttr"`
	EmailAttr       string `gorm:"size:64" json:"emailAttr"`
	// IDAttr is the identity attribute, one of ldap.IdentityAttributes.
	IDAttr string `gorm:"size:64;not null;default:entryUUID" json:"idAttr"`
	// DepartmentAttr is the department attribute; empty means not synced.
	DepartmentAttr string `gorm:"size:64;default:departmentNumber" json:"departmentAttr"`
	// Version increases on every save. A login records it before the
	// password check and refuses to write back if it changed meanwhile.
	Version uint64 `gorm:"not null;default:0" json:"-"`
}

func (SystemLDAPConfig) TableName() string {
	return "sys_ldap_config"
}

type SystemAuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ActorUserID   uint      `gorm:"index" json:"actorUserId"`
	ActorUsername string    `gorm:"size:128" json:"actorUsername"`
	Action        string    `gorm:"size:128;not null;index" json:"action"`
	ResourceType  string    `gorm:"size:128;index" json:"resourceType"`
	ResourceID    string    `gorm:"size:128;index" json:"resourceId"`
	Result        string    `gorm:"size:32;not null;index" json:"result"`
	IP            string    `gorm:"size:64" json:"ip"`
	UserAgent     string    `gorm:"size:255" json:"userAgent"`
	Detail        string    `gorm:"type:text" json:"detail"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (SystemAuditLog) TableName() string {
	return "sys_audit_logs"
}

// SystemDataMigration records a one-time data migration that has completed.
// The row is written in the same transaction as the migration's changes, so a
// migration that failed or was interrupted runs again on the next start.
type SystemDataMigration struct {
	Name      string    `gorm:"primaryKey;size:128" json:"name"`
	AppliedAt time.Time `gorm:"not null" json:"appliedAt"`
}

func (SystemDataMigration) TableName() string {
	return "sys_data_migrations"
}
