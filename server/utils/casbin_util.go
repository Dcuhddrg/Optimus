package utils

import (
	"errors"
	"sync"
	"sync/atomic"

	"optimus/server/global"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

const casbinModelText = `
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

var (
	currentEnforcer atomic.Pointer[casbin.SyncedEnforcer]
	// policyStale is set when the latest reload failed: the database may hold
	// policies the in-memory enforcer does not, so it must not authorize.
	policyStale atomic.Bool
	reloadMu    sync.Mutex

	// BuildCasbinEnforcer loads the complete policy from casbin_rule into a new
	// enforcer. Tests replace it to inject failures or delays.
	BuildCasbinEnforcer = buildCasbinEnforcer
)

// GetCasbin returns the process-wide enforcer over the casbin_rule table, as
// GVA's utils.GetCasbin, loading it on first use. Policies are
// p = <roleID>, <path without router prefix>, <method>; paths match with
// keyMatch2 (":id" segments). It returns nil while no policy was ever loaded.
// Unlike GVA it has no decision cache: a decision computed from the old policy
// could otherwise be written back after a reload and outlive it (design 3.1).
func GetCasbin() *casbin.SyncedEnforcer {
	if e := currentEnforcer.Load(); e != nil {
		return e
	}
	if !policyStale.Load() {
		_ = ReloadCasbin()
	}
	return currentEnforcer.Load()
}

// CasbinPolicyStale reports whether authorization must fail closed because
// the latest reload failed or no policy was ever loaded.
func CasbinPolicyStale() bool {
	return policyStale.Load() || currentEnforcer.Load() == nil
}

// ReloadCasbin builds a fresh enforcer from the database and swaps it in.
// Reloads are serialized and read the database only after taking the lock, so
// a reload that finishes later always saw later commits. On failure the old
// enforcer stays in place but is marked stale.
func ReloadCasbin() error {
	reloadMu.Lock()
	defer reloadMu.Unlock()
	e, err := BuildCasbinEnforcer()
	if err != nil {
		policyStale.Store(true)
		return err
	}
	currentEnforcer.Store(e)
	policyStale.Store(false)
	return nil
}

func buildCasbinEnforcer() (*casbin.SyncedEnforcer, error) {
	if global.OPTIMUS_DB == nil {
		return nil, errors.New("database is not initialized")
	}
	a, err := gormadapter.NewAdapterByDB(global.OPTIMUS_DB)
	if err != nil {
		return nil, err
	}
	m, err := model.NewModelFromString(casbinModelText)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewSyncedEnforcer(m, a)
	if err != nil {
		return nil, err
	}
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	return e, nil
}

// ResetCasbin drops the enforcer so the next GetCasbin binds to the current
// global.OPTIMUS_DB. GVA has no equivalent; tests swap databases.
func ResetCasbin() {
	reloadMu.Lock()
	defer reloadMu.Unlock()
	currentEnforcer.Store(nil)
	policyStale.Store(false)
}
