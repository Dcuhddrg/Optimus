package cloud

import "time"

// AccountView is the API-safe representation (credential masked).
type AccountView struct {
	ID             uint              `json:"id"`
	ProjectID      uint              `json:"projectId"`
	Name           string            `json:"name"`
	Provider       string            `json:"provider"`
	AccountID      string            `json:"accountId"`
	CredentialType string            `json:"credentialType"`
	Credential     string            `json:"credential"`
	Tags           map[string]string `json:"tags"`
	Status         string            `json:"status"`
	LastSyncedAt   *time.Time        `json:"lastSyncedAt"`
}

// CreateAccountRequest is the body for POST /cloud/accounts.
type CreateAccountRequest struct {
	Name           string            `json:"name" binding:"required"`
	Provider       string            `json:"provider" binding:"required"`
	AccountID      string            `json:"accountId" binding:"required"`
	CredentialType string            `json:"credentialType" binding:"required"`
	Credential     map[string]string `json:"credential" binding:"required"`
	Tags           map[string]string `json:"tags"`
}

// UpdateAccountTagsRequest is the body for PUT /cloud/accounts/:id/tags.
type UpdateAccountTagsRequest struct {
	Tags map[string]string `json:"tags" binding:"required"`
}

// UpdateAccountRequest is the body for PUT /cloud/accounts/:id.
type UpdateAccountRequest struct {
	Name       *string           `json:"name"`
	Credential map[string]string `json:"credential"`
}

// AssetPage is the paginated response for cloud asset queries.
type AssetPage struct {
	Items    []AssetView `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// AssetView is the API representation of a single cloud asset.
type AssetView struct {
	ID           uint       `json:"id"`
	AccountID    uint       `json:"accountId"`
	Provider     string     `json:"provider"`
	Region       string     `json:"region"`
	ResourceType string     `json:"resourceType"`
	ResourceID   string     `json:"resourceId"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	ConfigJSON   string     `json:"configJson"`
	TagsJSON     string     `json:"tagsJson"`
	DiscoveredAt time.Time  `json:"discoveredAt"`
	LastSeenAt   time.Time  `json:"lastSeenAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
}

// SyncResult reports the outcome of an account sync.
type SyncResult struct {
	Upserted int `json:"upserted"`
	Errors   int `json:"errors"`
}
