package response

import (
	"time"
)

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
