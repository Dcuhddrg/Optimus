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

func SeedCloudData(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if err := seedCloudPermissions(db); err != nil {
		return err
	}
	return seedAdminCloudPermissions(db)
}

// DefaultCloudPermissions returns all cloud API permissions.
func DefaultCloudPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "cloud.accounts.view", Name: "View cloud accounts", Type: "api", Method: "GET", Path: "/api/v1/cloud/accounts", Sort: 1000},
		{Code: "cloud.accounts.detail", Name: "View cloud account detail", Type: "api", Method: "GET", Path: "/api/v1/cloud/accounts/:id", Sort: 1001},
		{Code: "cloud.accounts.create", Name: "Create cloud accounts", Type: "api", Method: "POST", Path: "/api/v1/cloud/accounts", Sort: 1002},
		{Code: "cloud.accounts.update", Name: "Update cloud accounts", Type: "api", Method: "PUT", Path: "/api/v1/cloud/accounts/:id", Sort: 1003},
		{Code: "cloud.accounts.delete", Name: "Delete cloud accounts", Type: "api", Method: "DELETE", Path: "/api/v1/cloud/accounts/:id", Sort: 1004},
		{Code: "cloud.accounts.sync", Name: "Sync cloud accounts", Type: "api", Method: "POST", Path: "/api/v1/cloud/accounts/:id/sync", Sort: 1005},
		{Code: "cloud.assets.view", Name: "View cloud assets", Type: "api", Method: "GET", Path: "/api/v1/cloud/assets", Sort: 1100},
	}
}

func seedCloudPermissions(db *gorm.DB) error {
	for _, p := range DefaultCloudPermissions() {
		if err := db.Where("code = ?", p.Code).FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedAdminCloudPermissions(db *gorm.DB) error {
	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}
	for _, p := range DefaultCloudPermissions() {
		var perm systemModel.SystemPermission
		if err := db.Where("code = ?", p.Code).First(&perm).Error; err != nil {
			return err
		}
		rp := systemModel.SystemRolePermission{RoleID: adminRole.ID, PermissionID: perm.ID}
		if err := db.Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).FirstOrCreate(&rp).Error; err != nil {
			return err
		}
	}
	return nil
}
