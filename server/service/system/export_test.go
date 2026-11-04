package system

import (
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	ldapUtil "optimus/server/utils/ldap"
)

// NewLDAPServiceForTest makes LDAP calls use client until the test ends.
func NewLDAPServiceForTest(t testing.TB, client ldapUtil.Client) *LDAPService {
	t.Helper()
	saved := newLDAPClient
	t.Cleanup(func() { newLDAPClient = saved })
	newLDAPClient = func() ldapUtil.Client { return client }
	return &LDAPService{}
}

// NewAuthServiceForTest sets a signing config so login can issue tokens.
func NewAuthServiceForTest(t testing.TB) *AuthService {
	t.Helper()
	saved := global.OPTIMUS_CONFIG.JWT
	t.Cleanup(func() { global.OPTIMUS_CONFIG.JWT = saved })
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "test-signing-key", ExpiresTime: 3600, Issuer: "optimus-test"}
	return &AuthService{}
}
