// Package system seeds the system module, as GVA's source/system.
package system

import (
	"context"
	"errors"
	"strings"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"gorm.io/gorm"
)

const (
	initOrderSystem      = 100
	defaultAdminPassword = "admin"
)

type initSystem struct{}

func init() {
	systemService.RegisterInit(initOrderSystem, &initSystem{})
}

func (i *initSystem) InitializerName() string { return "system" }

func (i *initSystem) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	return ctx, systemService.AutoMigrateSystem(db)
}

func (i *initSystem) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&systemModel.SystemUser{})
}

// DataInserted is false so idempotent seeds and upgrades run on every start.
func (i *initSystem) DataInserted(ctx context.Context) bool { return false }

// InitializeData seeds roles, permissions and the admin account, and upgrades
// legacy API grants to Casbin policies.
func (i *initSystem) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	password, _ := ctx.Value("adminPassword").(string)
	return ctx, SeedSystemData(db, password)
}

// SeedSystemData is idempotent: it creates missing rows and never overwrites edits.
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
	if err := migrateAPIGrantsToCasbin(db); err != nil {
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

func seedRoles(db *gorm.DB) error {
	for _, role := range systemService.DefaultSystemRoles() {
		if err := db.Where("code = ?", role.Code).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPermissions(db *gorm.DB) error {
	for _, permission := range systemService.DefaultSystemPermissions() {
		if err := db.Where("code = ?", permission.Code).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedAdminUser(db *gorm.DB, initialAdminPassword string) error {
	var existing systemModel.SystemUser
	err := db.Where("username = ?", systemService.DefaultAdminUsername).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := utils.BcryptHash(initialAdminPassword)
	if err != nil {
		return err
	}

	admin := systemModel.SystemUser{
		UUID:         "00000000-0000-0000-0000-000000000001",
		Username:     systemService.DefaultAdminUsername,
		DisplayName:  "Administrator",
		Source:       "local",
		Status:       systemService.DefaultUserStatus,
		PasswordHash: hash,
	}
	return db.Create(&admin).Error
}

func seedAdminUserRole(db *gorm.DB) error {
	var admin systemModel.SystemUser
	if err := db.Where("username = ?", systemService.DefaultAdminUsername).First(&admin).Error; err != nil {
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

	return systemService.GrantPermissions(db, adminRole.ID, permissions)
}

// migrateAPIGrantsToCasbin upgrades databases written before #56: API catalog
// paths lose the /api/v1 prefix, and API grants move from sys_role_permissions
// to casbin_rule. Safe to run on every start.
func migrateAPIGrantsToCasbin(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var prefixed []systemModel.SystemPermission
		if err := tx.Where("type = ? AND path LIKE ?", "api", utils.DefaultRouterPrefix+"/%").Find(&prefixed).Error; err != nil {
			return err
		}
		for _, permission := range prefixed {
			path := strings.TrimPrefix(permission.Path, utils.DefaultRouterPrefix)
			if err := tx.Model(&systemModel.SystemPermission{}).Where("id = ?", permission.ID).Update("path", path).Error; err != nil {
				return err
			}
		}
		var grants []struct {
			RoleID uint
			ID     uint
			Path   string
			Method string
		}
		if err := tx.Table("sys_role_permissions AS rp").
			Select("rp.role_id, rp.id, p.path, p.method").
			Joins("JOIN sys_permissions AS p ON p.id = rp.permission_id").
			Where("p.type = ?", "api").Scan(&grants).Error; err != nil {
			return err
		}
		for _, grant := range grants {
			if err := systemService.CasbinServiceApp.AddPolicies(tx, grant.RoleID, []systemService.CasbinInfo{{Path: grant.Path, Method: grant.Method}}); err != nil {
				return err
			}
			if err := tx.Delete(&systemModel.SystemRolePermission{}, grant.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
