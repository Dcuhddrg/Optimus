package utils

import (
	"sync"
	"time"

	"optimus/server/global"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
)

var (
	syncedCachedEnforcer *casbin.SyncedCachedEnforcer
	once                 sync.Once
)

// GetCasbin lazily builds the process-wide enforcer over the casbin_rule table,
// as GVA's utils.GetCasbin. Policies are p = <roleID>, <path without router
// prefix>, <method>; paths match with keyMatch2 (":id" segments).
func GetCasbin() *casbin.SyncedCachedEnforcer {
	once.Do(func() {
		a, err := gormadapter.NewAdapterByDB(global.OPTIMUS_DB)
		if err != nil {
			zap.L().Error("casbin adapter init failed; check casbin_rule table", zap.Error(err))
			return
		}
		text := `
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[role_definition]
		g = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = r.sub == p.sub && keyMatch2(r.obj,p.obj) && r.act == p.act
		`
		m, err := model.NewModelFromString(text)
		if err != nil {
			zap.L().Error("casbin model load failed", zap.Error(err))
			return
		}
		enforcer, err := casbin.NewSyncedCachedEnforcer(m, a)
		if err != nil {
			zap.L().Error("casbin enforcer init failed", zap.Error(err))
			return
		}
		enforcer.SetExpireTime(time.Hour)
		if err = enforcer.LoadPolicy(); err != nil {
			zap.L().Error("casbin policy load failed", zap.Error(err))
			return
		}
		syncedCachedEnforcer = enforcer
	})
	return syncedCachedEnforcer
}

// ResetCasbin drops the cached enforcer so the next GetCasbin binds to the
// current global.OPTIMUS_DB. GVA has no equivalent; tests swap databases.
func ResetCasbin() {
	syncedCachedEnforcer = nil
	once = sync.Once{}
}
