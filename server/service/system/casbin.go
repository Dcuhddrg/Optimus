package system

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"optimus/server/global"

	systemModel "optimus/server/model/system"
	"optimus/server/utils"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
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

// ErrCasbinReloadPending means the policy change is committed but the
// in-memory enforcer could not be reloaded yet. Authorization fails closed
// until a background retry succeeds.
var ErrCasbinReloadPending = errors.New("permissions saved; policy reload pending")

var (
	casbinRetrying     atomic.Bool
	casbinRetryInitial = time.Second
	casbinRetryMax     = 30 * time.Second
)

// FreshCasbin reloads all policies from casbin_rule after a committed change,
// as GVA. On failure it starts a background retry and returns
// ErrCasbinReloadPending; requests needing API permissions are rejected until
// the retry succeeds (design 3.1).
func (casbinService *CasbinService) FreshCasbin() error {
	if err := utils.ReloadCasbin(); err != nil {
		logCasbinError("casbin policy reload failed; failing closed until retry succeeds", err)
		startCasbinRetry()
		return fmt.Errorf("%w: %v", ErrCasbinReloadPending, err)
	}
	return nil
}

// Enforcer returns the enforcer to authorize with, or false while the policy
// is stale. A stale or never-loaded policy also (re)starts the retry loop.
func (casbinService *CasbinService) Enforcer() (*casbin.SyncedEnforcer, bool) {
	// GetCasbin loads the policy on first use; a failed load marks it stale.
	if e := utils.GetCasbin(); e != nil && !utils.CasbinPolicyStale() {
		return e, true
	}
	startCasbinRetry()
	return nil, false
}

func startCasbinRetry() {
	if !casbinRetrying.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer casbinRetrying.Store(false)
		delay := casbinRetryInitial
		for {
			time.Sleep(delay)
			if err := utils.ReloadCasbin(); err == nil {
				InvalidateAllAuthState()
				if logger := global.OPTIMUS_LOG; logger != nil {
					logger.Info("casbin policy reload recovered")
				}
				return
			}
			delay *= 2
			if delay > casbinRetryMax {
				delay = casbinRetryMax
			}
		}
	}()
}

func logCasbinError(msg string, err error) {
	if logger := global.OPTIMUS_LOG; logger != nil {
		logger.Error(msg, zap.Error(err))
		return
	}
	log.Printf("%s: %v", msg, err)
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
