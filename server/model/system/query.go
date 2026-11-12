package system

import "time"

type SystemRoleSummary struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	// Approvable lets authorization drop approval roles for external
	// accounts (design 3.6).
	Approvable bool `json:"approvable"`
}

type AuditPage struct {
	Items    []AuditLogEntry `json:"items"`
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
