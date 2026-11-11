package system

import "sync"

// ldapWriteBack serializes every "read from LDAP, write back to the database"
// step (design 3.3, write ordering): each write uses data read while holding
// the lock, so a later write is always based on a later read. One global
// lock, not per user: at login the identity is not known yet, legacy accounts
// have none, and one account can move between identities.
//
// In-process only, enough for a single instance; a multi-instance deployment
// needs a database or distributed lock, together with the auth state cache.
var ldapWriteBack sync.Mutex
