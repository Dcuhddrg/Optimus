package system

import (
	"context"
	"errors"
	"sync"

	"optimus/server/global"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

// UserAuthState is what each authenticated request is authorized against:
// loaded per request (through the cache) instead of trusting the token, so a
// disabled user or a revoked role takes effect on the next request.
type UserAuthState struct {
	Exists  bool
	Status  string
	RoleIDs []uint
}

// Enabled reports whether the user may use the system at all.
func (s UserAuthState) Enabled() bool {
	return s.Exists && s.Status == DefaultUserStatus
}

// authStateCache keeps UserAuthState per user ID. Every invalidation bumps a
// global version under the same mutex that guards "compare version and store",
// so a loader that read the database before an invalidation can never write
// its stale result back afterwards (design 3.1).
type authStateCache struct {
	mu      sync.RWMutex
	version uint64
	entries map[uint]UserAuthState
}

func newAuthStateCache() *authStateCache {
	return &authStateCache{entries: map[uint]UserAuthState{}}
}

func (c *authStateCache) get(userID uint) (UserAuthState, uint64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state, ok := c.entries[userID]
	return state, c.version, ok
}

// storeIfVersion stores state only if no invalidation happened since the
// loader observed version.
func (c *authStateCache) storeIfVersion(userID uint, state UserAuthState, version uint64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.version != version {
		return false
	}
	if hook := authStateTestHooks.insideStore; hook != nil {
		hook(userID)
	}
	c.entries[userID] = state
	return true
}

func (c *authStateCache) invalidateUser(userID uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.version++
	delete(c.entries, userID)
}

func (c *authStateCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.version++
	c.entries = map[uint]UserAuthState{}
}

var authStates = newAuthStateCache()

// authStateTestHooks inject pause points for the interleaving tests:
// afterLoad runs after the database read and before the store attempt;
// insideStore runs inside the store lock after the version check passed.
var authStateTestHooks struct {
	afterLoad   func(userID uint)
	insideStore func(userID uint)
}

// InvalidateUserAuthState must be called after committing any write that
// changes this user's status or roles, before the write API returns.
func InvalidateUserAuthState(userID uint) { authStates.invalidateUser(userID) }

// InvalidateAllAuthState must be called after committing writes that can
// change many users' effective authorization (role deletion, role
// permissions, LDAP sync execution), before the write API returns.
func InvalidateAllAuthState() { authStates.invalidateAll() }

type AuthStateService struct{}

// LoadUserAuthState returns the user's current status and active role IDs.
func (s *AuthStateService) LoadUserAuthState(ctx context.Context, userID uint) (UserAuthState, error) {
	// The version observed before reading the database decides whether the
	// result may be cached.
	state, version, ok := authStates.get(userID)
	if ok {
		return state, nil
	}
	state, err := loadUserAuthStateFromDB(ctx, userID)
	if err != nil {
		return UserAuthState{}, err
	}
	if hook := authStateTestHooks.afterLoad; hook != nil {
		hook(userID)
	}
	authStates.storeIfVersion(userID, state, version)
	return state, nil
}

func loadUserAuthStateFromDB(ctx context.Context, userID uint) (UserAuthState, error) {
	if global.OPTIMUS_DB == nil {
		return UserAuthState{}, errors.New("database is not available")
	}
	var user systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Select("id", "status").Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UserAuthState{Exists: false, RoleIDs: []uint{}}, nil
	}
	if err != nil {
		return UserAuthState{}, err
	}
	roles, err := queryUserRoles(ctx, userID)
	if err != nil {
		return UserAuthState{}, err
	}
	roleIDs, _ := summarizeRoles(roles)
	return UserAuthState{Exists: true, Status: user.Status, RoleIDs: roleIDs}, nil
}
