package cloud

import (
	"time"

	"gorm.io/gorm"
)

type CloudAccount struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProjectID      uint           `gorm:"not null;default:0;index" json:"projectId"`
	Name           string         `gorm:"size:128;not null" json:"name"`
	Provider       string         `gorm:"size:32;not null;index" json:"provider"`
	AccountID      string         `gorm:"size:64;not null" json:"accountId"`
	CredentialType string         `gorm:"size:32;not null" json:"credentialType"`
	CredentialJSON string         `gorm:"type:text" json:"-"`
	Status         string         `gorm:"size:16;not null;index;default:'active'" json:"status"`
	LastSyncedAt   *time.Time     `json:"lastSyncedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

func (CloudAccount) TableName() string {
	return "cloud_accounts"
}

type CloudAsset struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	AccountID    uint       `gorm:"not null;index:idx_cloud_asset_account_type;uniqueIndex:uniq_cloud_asset_account_resource" json:"accountId"`
	Provider     string     `gorm:"size:32;not null;index:idx_cloud_asset_provider_region_type" json:"provider"`
	Region       string     `gorm:"size:64;not null;index:idx_cloud_asset_provider_region_type" json:"region"`
	ResourceType string     `gorm:"size:64;not null;index:idx_cloud_asset_account_type;index:idx_cloud_asset_provider_region_type" json:"resourceType"`
	ResourceID   string     `gorm:"size:128;not null;uniqueIndex:uniq_cloud_asset_account_resource" json:"resourceId"`
	Name         string     `gorm:"size:256" json:"name"`
	Status       string     `gorm:"size:32;not null;default:''" json:"status"`
	ConfigJSON   string     `gorm:"type:text" json:"configJson"`
	TagsJSON     string     `gorm:"type:text" json:"tagsJson"`
	DiscoveredAt time.Time  `json:"discoveredAt"`
	LastSeenAt   time.Time  `json:"lastSeenAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (CloudAsset) TableName() string {
	return "cloud_assets"
}
