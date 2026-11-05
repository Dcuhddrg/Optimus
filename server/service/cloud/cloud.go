package cloud

import (
	"errors"

	cloudModel "optimus/server/model/cloud"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

func AutoMigrateCloud(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	return db.AutoMigrate(
		&cloudModel.CloudAccount{},
		&cloudModel.CloudAsset{},
	)
}

// DefaultCloudPermissions returns all cloud menu and API permissions.
func DefaultCloudPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "cloud.accounts.menu", Name: "Cloud Accounts", Type: "menu", Path: "/cloud/accounts", Sort: 31},
		{Code: "cloud.assets.menu", Name: "Cloud Assets", Type: "menu", Path: "/cloud/assets", Sort: 32},
		{Code: "cloud.accounts.view", Name: "View cloud accounts", Type: "api", Method: "GET", Path: "/cloud/accounts", Sort: 1000},
		{Code: "cloud.accounts.detail", Name: "View cloud account detail", Type: "api", Method: "GET", Path: "/cloud/accounts/:id", Sort: 1001},
		{Code: "cloud.accounts.create", Name: "Create cloud accounts", Type: "api", Method: "POST", Path: "/cloud/accounts", Sort: 1002},
		{Code: "cloud.accounts.update", Name: "Update cloud accounts", Type: "api", Method: "PUT", Path: "/cloud/accounts/:id", Sort: 1003},
		{Code: "cloud.accounts.delete", Name: "Delete cloud accounts", Type: "api", Method: "DELETE", Path: "/cloud/accounts/:id", Sort: 1004},
		{Code: "cloud.accounts.sync", Name: "Sync cloud accounts", Type: "api", Method: "POST", Path: "/cloud/accounts/:id/sync", Sort: 1005},
		{Code: "cloud.accounts.tags.update", Name: "Update cloud account tags", Type: "api", Method: "PUT", Path: "/cloud/accounts/:id/tags", Sort: 1006},
		{Code: "cloud.assets.view", Name: "View cloud assets", Type: "api", Method: "GET", Path: "/cloud/assets", Sort: 1100},
	}
}
