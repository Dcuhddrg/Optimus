package system

import "time"

type LoginRequest struct {
	Username string
	Password string
}

type LoginResult struct {
	Token           string            `json:"token"`
	User            SystemUserSummary `json:"user"`
	Roles           []string          `json:"roles"`
	MenuPermissions []string          `json:"menuPermissions"`
}

type CurrentUserResult struct {
	User            SystemUserSummary `json:"user"`
	Roles           []string          `json:"roles"`
	MenuPermissions []string          `json:"menuPermissions"`
}

type SystemUserSummary struct {
	ID          uint   `json:"id"`
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Source      string `json:"source"`
	Status      string `json:"status"`
}

type SystemRoleSummary struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
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

type RoleRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleListResult struct {
	Items []SystemRole `json:"items"`
	Total int          `json:"total"`
}

type LDAPConfigRequest struct {
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	BaseDN          string `json:"baseDn"`
	BindDN          string `json:"bindDn"`
	BindPassword    string `json:"bindPassword"`
	UserFilter      string `json:"userFilter"`
	UsernameAttr    string `json:"usernameAttr"`
	DisplayNameAttr string `json:"displayNameAttr"`
	EmailAttr       string `json:"emailAttr"`
}

type LDAPConfigResult struct {
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	BaseDN          string `json:"baseDn"`
	BindDN          string `json:"bindDn"`
	BindPassword    string `json:"bindPassword,omitempty"`
	HasBindPassword bool   `json:"hasBindPassword"`
	UserFilter      string `json:"userFilter"`
	UsernameAttr    string `json:"usernameAttr"`
	DisplayNameAttr string `json:"displayNameAttr"`
	EmailAttr       string `json:"emailAttr"`
}

type LDAPTestRequest struct {
	URL                  string `json:"url"`
	BaseDN               string `json:"baseDn"`
	BindDN               string `json:"bindDn"`
	BindPassword         string `json:"bindPassword"`
	UseSavedBindPassword bool   `json:"useSavedBindPassword"`
	UserFilter           string `json:"userFilter"`
	UsernameAttr         string `json:"usernameAttr"`
	DisplayNameAttr      string `json:"displayNameAttr"`
	EmailAttr            string `json:"emailAttr"`
}

type LDAPTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AuditPage struct {
	Items    []AuditLogEntry `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// AuditListData is the HTTP data payload for migrated audit queries.
// AuditPage remains the internal Service/Store contract.
type AuditListData struct {
	List     []AuditLogEntry `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type AuditLogEntry struct {
	ID            uint      `json:"id"`
	ActorUserID   uint      `json:"actorUserId"`
	ActorUsername string    `json:"actorUsername"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resourceType"`
	ResourceID    string    `json:"resourceId"`
	Result        string    `json:"result"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"userAgent"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"createdAt"`
}
