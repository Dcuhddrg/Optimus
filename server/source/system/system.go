// Package system seeds the system module, as GVA's source/system.
package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return ctx, MigrateSystemSchema(db)
}

// MigrateSystemSchema migrates the system tables and runs the one-time data
// migrations.
func MigrateSystemSchema(db *gorm.DB) error {
	if err := systemService.AutoMigrateSystem(db); err != nil {
		return err
	}
	if err := runDataMigration(db, presetApprovalRolesMigration, adoptPresetApprovalRoles); err != nil {
		return err
	}
	return MigrateRemoveRoleStatus(db)
}

const presetApprovalRolesMigration = "2026-11-preset-approval-roles"

// runDataMigration runs fn once. fn and the completion marker share one
// transaction: a failure or crash leaves no marker, so the next start retries.
func runDataMigration(db *gorm.DB, name string, fn func(tx *gorm.DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var done int64
		if err := tx.Model(&systemModel.SystemDataMigration{}).Where("name = ?", name).Count(&done).Error; err != nil {
			return err
		}
		if done > 0 {
			return nil
		}
		if err := fn(tx); err != nil {
			return fmt.Errorf("data migration %s: %w", name, err)
		}
		return tx.Create(&systemModel.SystemDataMigration{Name: name, AppliedAt: time.Now()}).Error
	})
}

// presetAdoptTestHook, when set, runs after each preset role is handled so
// tests can fail the migration midway.
var presetAdoptTestHook func() error

// adoptPresetApprovalRoles handles roles that already use a preset approval
// code when the preset roles are introduced. Marking a role approvable makes
// its holders approval candidates, which nobody authorized, so only a role
// with no holders and no permissions is adopted. Any other same-code role is
// left untouched and audited for the admin to decide in role management.
// Later edits to the flag are never overwritten, since this runs once.
func adoptPresetApprovalRoles(db *gorm.DB) error {
	for _, preset := range systemService.DefaultSystemRoles() {
		if !preset.Approvable {
			continue
		}
		var role systemModel.SystemRole
		err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", preset.Code).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if role.Approvable {
			continue
		}
		var menuGrants, policies, holders int64
		if err := db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ?", role.ID).Count(&menuGrants).Error; err != nil {
			return err
		}
		if err := db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ?", "p", strconv.Itoa(int(role.ID))).Count(&policies).Error; err != nil {
			return err
		}
		if err := db.Model(&systemModel.SystemUserRole{}).Where("role_id = ?", role.ID).Count(&holders).Error; err != nil {
			return err
		}
		adopt := menuGrants+policies == 0 && holders == 0
		reason, result := "preset_role_adopted", "success"
		if adopt {
			if err := db.Model(&systemModel.SystemRole{}).Where("id = ?", role.ID).Update("approvable", true).Error; err != nil {
				return err
			}
		} else {
			reason, result = "preset_role_conflict", "failure"
		}
		detail, _ := json.Marshal(map[string]any{"version": 1, "reason": reason, "roleCode": role.Code, "permissions": menuGrants + policies, "holders": holders})
		if err := db.Create(&systemModel.SystemAuditLog{
			Action:       "role.preset.adopt",
			ResourceType: "role",
			ResourceID:   strconv.Itoa(int(role.ID)),
			Result:       result,
			Detail:       string(detail),
		}).Error; err != nil {
			return err
		}
		fields := []zap.Field{zap.String("role", role.Code), zap.Int64("permissions", menuGrants+policies), zap.Int64("holders", holders)}
		if adopt {
			zap.L().Info("existing empty role adopted as preset approval role", fields...)
		} else {
			zap.L().Warn("existing role uses a preset approval code but has holders or permissions; left non-approvable, mark it approvable in role management if intended", fields...)
		}
		if presetAdoptTestHook != nil {
			if err := presetAdoptTestHook(); err != nil {
				return err
			}
		}
	}
	return nil
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
	if err := removeRetiredPermissions(db); err != nil {
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

// MigrateRemoveRoleStatus drops the role status column, as GVA has none
// (design 3.2). Users first lose their bindings to disabled roles, so every
// user keeps exactly the permissions they effectively had: a disabled role
// granted nothing. Each unbinding is audited and logged. It runs before
// serving and is idempotent: a crash before the column drop leaves disabled
// roles without holders and the next start finishes the job.
func MigrateRemoveRoleStatus(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&systemModel.SystemRole{}, "status") {
		return nil
	}
	var disabled []struct {
		ID   uint
		Code string
	}
	if err := db.Table("sys_roles").Select("id, code").Where("status <> ?", "enabled").Scan(&disabled).Error; err != nil {
		return err
	}
	for _, role := range disabled {
		var userIDs []uint
		if err := db.Model(&systemModel.SystemUserRole{}).Where("role_id = ?", role.ID).Order("user_id ASC").Pluck("user_id", &userIDs).Error; err != nil {
			return err
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, userID := range userIDs {
				detail, _ := json.Marshal(map[string]any{"version": 1, "reason": "disabled_role_unbound", "roleId": role.ID, "roleCode": role.Code})
				if err := tx.Create(&systemModel.SystemAuditLog{
					Action:       "user.roles.migrate",
					ResourceType: "user",
					ResourceID:   strconv.Itoa(int(userID)),
					Result:       "success",
					Detail:       string(detail),
				}).Error; err != nil {
					return err
				}
			}
			return tx.Where("role_id = ?", role.ID).Delete(&systemModel.SystemUserRole{}).Error
		}); err != nil {
			return err
		}
		zap.L().Info("role status migration: unbound users from disabled role", zap.String("role", role.Code), zap.Uints("users", userIDs))
	}
	if err := db.Exec("ALTER TABLE `sys_roles` DROP COLUMN `status`").Error; err != nil {
		return err
	}
	zap.L().Info("role status migration: dropped sys_roles.status", zap.Int("disabledRoles", len(disabled)))
	return nil
}

// removeRetiredPermissions deletes catalog entries the product no longer
// has, with their menu grants and Casbin policies.
func removeRetiredPermissions(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, retired := range systemService.RetiredSystemPermissions() {
			var ids []uint
			if err := tx.Model(&systemModel.SystemPermission{}).Where("code = ?", retired.Code).Pluck("id", &ids).Error; err != nil {
				return err
			}
			if len(ids) > 0 {
				if err := tx.Where("permission_id IN ?", ids).Delete(&systemModel.SystemRolePermission{}).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", ids).Delete(&systemModel.SystemPermission{}).Error; err != nil {
					return err
				}
			}
			if retired.Type == "api" {
				if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", retired.Path, retired.Method).Delete(&gormadapter.CasbinRule{}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
