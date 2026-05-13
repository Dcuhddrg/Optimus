package system

import "time"

type SystemUser struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UUID         string     `gorm:"size:64;uniqueIndex;not null" json:"uuid"`
	Username     string     `gorm:"size:128;uniqueIndex;not null" json:"username"`
	DisplayName  string     `gorm:"size:128" json:"displayName"`
	Email        string     `gorm:"size:255" json:"email"`
	Source       string     `gorm:"size:32;not null;index" json:"source"`
	Status       string     `gorm:"size:32;not null;index" json:"status"`
	PasswordHash string     `gorm:"size:255" json:"-"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (SystemUser) TableName() string {
	return "sys_users"
}

type SystemRole struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Status      string    `gorm:"size:32;not null;index" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
	ID         uint      `gorm:"primaryKey" json:"id"`
	Code       string    `gorm:"size:160;uniqueIndex;not null" json:"code"`
	Name       string    `gorm:"size:128;not null" json:"name"`
	Type       string    `gorm:"size:32;not null;index" json:"type"`
	Method     string    `gorm:"size:16;index" json:"method"`
	Path       string    `gorm:"size:255;index" json:"path"`
	ParentCode string    `gorm:"size:160;index" json:"parentCode"`
	Sort       int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
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
	ID              uint      `gorm:"primaryKey" json:"id"`
	Enabled         bool      `gorm:"not null;default:false" json:"enabled"`
	URL             string    `gorm:"size:255" json:"url"`
	BaseDN          string    `gorm:"size:255" json:"baseDn"`
	BindDN          string    `gorm:"size:255" json:"bindDn"`
	BindPassword    string    `gorm:"size:255" json:"-"`
	UserFilter      string    `gorm:"size:255" json:"userFilter"`
	UsernameAttr    string    `gorm:"size:64" json:"usernameAttr"`
	DisplayNameAttr string    `gorm:"size:64" json:"displayNameAttr"`
	EmailAttr       string    `gorm:"size:64" json:"emailAttr"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
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
