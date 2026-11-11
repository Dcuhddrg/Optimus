// Package system seeds the system module, as GVA's source/system.
package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
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
	// Normalize legacy /api/v1-prefixed catalog paths first: grants written
	// below copy the catalog path into casbin_rule, and requests are
	// authorized on the unprefixed path.
	if err := migrateAPIGrantsToCasbin(db); err != nil {
		return err
	}
	// Repairs are only reported for an existing admin; a fresh install
	// creates and grants everything without being a repair.
	repairs := []string{}
	// Catalog entries and their first grant commit together: if startup
	// fails in between, the next start still sees them as new and grants them.
	if err := db.Transaction(func(tx *gorm.DB) error {
		createdPermissions, err := systemService.SeedPermissionCatalog(tx, systemService.DefaultSystemPermissions())
		if err != nil {
			return err
		}
		return seedAdminRolePermissions(tx, createdPermissions, &repairs)
	}); err != nil {
		return err
	}
	created, err := seedAdminUser(db, initialAdminPassword, &repairs)
	if err != nil {
		return err
	}
	if created {
		repairs = nil
	}
	if err := seedAdminUserRole(db, &repairs); err != nil {
		return err
	}
	if created {
		return nil
	}
	return recordBuiltinAdminRepairs(db, repairs)
}

func seedRoles(db *gorm.DB) error {
	for _, role := range systemService.DefaultSystemRoles() {
		if err := db.Where("code = ?", role.Code).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedAdminUser creates the built-in admin, or re-enables it: it is the
// permanent fallback administrator, and rules before #80 allowed disabling it
// while other admins existed (design 3.4).
func seedAdminUser(db *gorm.DB, initialAdminPassword string, repairs *[]string) (created bool, err error) {
	var existing systemModel.SystemUser
	err = db.Where("uuid = ?", systemService.BuiltinAdminUUID).First(&existing).Error
	if err == nil {
		if existing.Status != systemService.DefaultUserStatus {
			if err := db.Model(&systemModel.SystemUser{}).Where("id = ?", existing.ID).Update("status", systemService.DefaultUserStatus).Error; err != nil {
				return false, err
			}
			*repairs = append(*repairs, "re-enabled")
		}
		return false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	// A row named admin with another UUID can only come from manual edits.
	// A local one is the built-in admin: adopt it by giving it the fixed UUID
	// (UUIDs are never used to look users up, so its sessions stay valid).
	// Anything else would shadow the built-in admin: refuse to start.
	if err := db.Where("username = ?", systemService.DefaultAdminUsername).First(&existing).Error; err == nil {
		if existing.Source != "local" {
			return false, fmt.Errorf("account %q (id %d) has source %q and blocks the built-in admin; rename it, e.g. UPDATE sys_users SET username = 'admin-%s' WHERE id = %d, then restart", existing.Username, existing.ID, existing.Source, existing.Source, existing.ID)
		}
		columns := map[string]any{"uuid": systemService.BuiltinAdminUUID}
		*repairs = append(*repairs, "adopted legacy local admin (uuid "+existing.UUID+")")
		if existing.Status != systemService.DefaultUserStatus {
			columns["status"] = systemService.DefaultUserStatus
			*repairs = append(*repairs, "re-enabled")
		}
		return false, db.Model(&systemModel.SystemUser{}).Where("id = ?", existing.ID).Updates(columns).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	hash, err := utils.BcryptHash(initialAdminPassword)
	if err != nil {
		return false, err
	}

	admin := systemModel.SystemUser{
		UUID:         systemService.BuiltinAdminUUID,
		Username:     systemService.DefaultAdminUsername,
		DisplayName:  "Administrator",
		Source:       "local",
		Status:       systemService.DefaultUserStatus,
		PasswordHash: hash,
	}
	return true, db.Create(&admin).Error
}

func seedAdminUserRole(db *gorm.DB, repairs *[]string) error {
	var admin systemModel.SystemUser
	if err := db.Where("uuid = ?", systemService.BuiltinAdminUUID).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", systemService.AdminRoleCode).First(&adminRole).Error; err != nil {
		return err
	}

	userRole := systemModel.SystemUserRole{UserID: admin.ID, RoleID: adminRole.ID}
	result := db.Where("user_id = ? AND role_id = ?", admin.ID, adminRole.ID).FirstOrCreate(&userRole)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		*repairs = append(*repairs, "admin role restored")
	}
	return nil
}

// seedAdminRolePermissions keeps every system.* permission on the admin role
// (missing ones are reported as repairs) and grants other permissions only
// when this run introduced them: permissions an administrator revoked from
// the admin role stay revoked across restarts (design 3.4).
func seedAdminRolePermissions(db *gorm.DB, created []systemModel.SystemPermission, repairs *[]string) error {
	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", systemService.AdminRoleCode).First(&adminRole).Error; err != nil {
		return err
	}

	var systemPermissions []systemModel.SystemPermission
	if err := db.Where("code LIKE ?", "system.%").Find(&systemPermissions).Error; err != nil {
		return err
	}
	missing, err := missingSystemGrants(db, adminRole.ID)
	if err != nil {
		return err
	}
	// A permission introduced by this run is granted for the first time, not
	// restored.
	isNew := map[string]bool{}
	for _, permission := range created {
		isNew[permission.Code] = true
	}
	restored := []string{}
	for _, code := range missing {
		if !isNew[code] {
			restored = append(restored, code)
		}
	}
	missing = restored
	if err := systemService.GrantPermissions(db, adminRole.ID, append(systemPermissions, created...)); err != nil {
		return err
	}
	if len(missing) > 0 {
		*repairs = append(*repairs, "system permissions restored: "+strings.Join(missing, ","))
	}
	return nil
}

// missingSystemGrants lists system.* permissions the role lacks: menus in
// sys_role_permissions, APIs in casbin_rule.
func missingSystemGrants(db *gorm.DB, roleID uint) ([]string, error) {
	var permissions []systemModel.SystemPermission
	if err := db.Where("code LIKE ?", "system.%").Order("sort ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	missing := []string{}
	for _, permission := range permissions {
		var count int64
		var err error
		if permission.Type == "api" {
			err = db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", strconv.Itoa(int(roleID)), permission.Path, permission.Method).Count(&count).Error
		} else {
			err = db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ? AND permission_id = ?", roleID, permission.ID).Count(&count).Error
		}
		if err != nil {
			return nil, err
		}
		if count == 0 {
			missing = append(missing, permission.Code)
		}
	}
	return missing, nil
}

// recordBuiltinAdminRepairs audits and loudly logs any repair of the built-in
// admin: if it had been disabled because of a weak password, that password
// works again, so operators must confirm it.
func recordBuiltinAdminRepairs(db *gorm.DB, repairs []string) error {
	if len(repairs) == 0 {
		return nil
	}
	var admin systemModel.SystemUser
	if err := db.Where("uuid = ?", systemService.BuiltinAdminUUID).First(&admin).Error; err != nil {
		return err
	}
	detail, _ := json.Marshal(map[string]any{"version": 1, "reason": "builtin_admin_repaired", "repairs": repairs})
	if err := db.Create(&systemModel.SystemAuditLog{
		Action:       "system.admin.repair",
		ResourceType: "user",
		ResourceID:   strconv.Itoa(int(admin.ID)),
		Result:       "success",
		Detail:       string(detail),
	}).Error; err != nil {
		return err
	}
	msg := "built-in admin repaired at startup"
	for _, repair := range repairs {
		if repair == "re-enabled" {
			msg += "; it had been disabled, confirm its password"
		}
	}
	zap.L().Warn(msg, zap.Strings("repairs", repairs))
	return nil
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
