package system

import (
	"strconv"

	systemModel "optimus/server/model/system"
	"optimus/server/utils"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// CasbinService keeps role API policies in casbin_rule, as GVA's CasbinService.
// Menu permissions stay in sys_role_permissions (GVA: sys_authority_menus).
type CasbinService struct{}

var CasbinServiceApp = new(CasbinService)

// CasbinInfo is one API policy: a path without router prefix and an HTTP method.
type CasbinInfo struct {
	Path   string
	Method string
}

// UpdateCasbin replaces all API policies of roleID within db (which may be a
// transaction). Call FreshCasbin after commit so the enforcer reloads.
func (casbinService *CasbinService) UpdateCasbin(db *gorm.DB, roleID uint, casbinInfos []CasbinInfo) error {
	sub := strconv.Itoa(int(roleID))
	if err := db.Where("ptype = ? AND v0 = ?", "p", sub).Delete(&gormadapter.CasbinRule{}).Error; err != nil {
		return err
	}
	return casbinService.AddPolicies(db, roleID, casbinInfos)
}

// AddPolicies grants API policies to roleID if they are not present yet.
func (casbinService *CasbinService) AddPolicies(db *gorm.DB, roleID uint, casbinInfos []CasbinInfo) error {
	sub := strconv.Itoa(int(roleID))
	for _, info := range casbinInfos {
		rule := gormadapter.CasbinRule{Ptype: "p", V0: sub, V1: info.Path, V2: info.Method}
		if err := db.Where(&rule).FirstOrCreate(&rule).Error; err != nil {
			return err
		}
	}
	return nil
}

// FreshCasbin reloads policies into the cached enforcer, as GVA.
func (casbinService *CasbinService) FreshCasbin() error {
	e := utils.GetCasbin()
	if e == nil {
		return nil
	}
	return e.LoadPolicy()
}

// APIPolicies maps API-type catalog permissions to Casbin policies.
func APIPolicies(permissions []systemModel.SystemPermission) []CasbinInfo {
	infos := make([]CasbinInfo, 0, len(permissions))
	for _, permission := range permissions {
		if permission.Type == "api" {
			infos = append(infos, CasbinInfo{Path: permission.Path, Method: permission.Method})
		}
	}
	return infos
}

// GrantPermissions gives roleID the listed permissions: menus through
// sys_role_permissions, APIs through casbin_rule. Used by module seeders.
func GrantPermissions(db *gorm.DB, roleID uint, permissions []systemModel.SystemPermission) error {
	for _, permission := range permissions {
		if permission.Type == "api" {
			continue
		}
		rp := systemModel.SystemRolePermission{RoleID: roleID, PermissionID: permission.ID}
		if err := db.Where("role_id = ? AND permission_id = ?", roleID, permission.ID).FirstOrCreate(&rp).Error; err != nil {
			return err
		}
	}
	return CasbinServiceApp.AddPolicies(db, roleID, APIPolicies(permissions))
}

// apiPermissionCodesForRole lists catalog codes of the API policies held by roleID.
func apiPermissionCodesForRole(db *gorm.DB, roleID uint) ([]string, error) {
	codes := make([]string, 0)
	err := db.Table("sys_permissions AS p").
		Joins("JOIN casbin_rule AS r ON r.ptype = 'p' AND r.v0 = ? AND r.v1 = p.path AND r.v2 = p.method", strconv.Itoa(int(roleID))).
		Where("p.type = ?", "api").Order("p.code").Pluck("p.code", &codes).Error
	return codes, err
}
