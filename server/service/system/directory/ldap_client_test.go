package directory

import (
	"testing"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/require"
)

func TestReplaceLDAPUsername(t *testing.T) {
	filter := "(&(objectClass=user)(sAMAccountName={username}))"
	result := replaceLDAPUsername(filter, "alice")
	require.Equal(t, "(&(objectClass=user)(sAMAccountName=alice))", result)
}

func TestReplaceLDAPUsernameEscapesSpecialChars(t *testing.T) {
	filter := "(uid={username})"
	result := replaceLDAPUsername(filter, "a*b")
	require.Equal(t, `(uid=a\2ab)`, result)
}

func TestFirstLDAPAttrReturnsValue(t *testing.T) {
	entry := &ldap.Entry{
		Attributes: []*ldap.EntryAttribute{
			{Name: "mail", Values: []string{"user@example.com"}},
		},
	}
	require.Equal(t, "user@example.com", firstLDAPAttr(entry, "mail", "fallback"))
}

func TestFirstLDAPAttrUsesFallbackWhenAttrEmpty(t *testing.T) {
	entry := &ldap.Entry{}
	require.Equal(t, "fallback", firstLDAPAttr(entry, "mail", "fallback"))
}

func TestFirstLDAPAttrUsesFallbackWhenAttrNameEmpty(t *testing.T) {
	entry := &ldap.Entry{
		Attributes: []*ldap.EntryAttribute{
			{Name: "mail", Values: []string{"user@example.com"}},
		},
	}
	require.Equal(t, "fallback", firstLDAPAttr(entry, "", "fallback"))
}
