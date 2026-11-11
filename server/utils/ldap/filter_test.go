package ldap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateLoginFilterDerivesSyncFilter(t *testing.T) {
	for _, tc := range []struct{ filter, sync string }{
		{"(uid={username})", "(uid=*)"},
		{"(UID={username})", "(uid=*)"},
		{"(&(uid={username})(objectClass=inetOrgPerson))", "(&(uid=*)(objectClass=inetOrgPerson))"},
		{"(&(objectClass=person)(uid={username})(memberOf=cn=ops,ou=groups,dc=optimus,dc=test)(!(employeeType=disabled)))", "(&(objectClass=person)(uid=*)(memberOf=cn=ops,ou=groups,dc=optimus,dc=test)(!(employeeType=disabled)))"},
	} {
		sync, err := ValidateLoginFilter(tc.filter, "uid")
		require.NoError(t, err, tc.filter)
		require.Equal(t, tc.sync, sync, tc.filter)
	}
}

func TestValidateLoginFilterRejectsAmbiguousStructures(t *testing.T) {
	for _, filter := range []string{
		"",
		"(objectClass=person)",
		"(&(uid={username})(!(cn={username})))",
		"(|(uid={username})(mail={username}))",
		"(|(uid={username})(objectClass=x))",
		"(!(uid={username}))",
		"(&(objectClass=x)(&(uid={username})))",
		"(uid=a{username})",
		"(uid={username}*)",
		"(cn={username})",
		"(&(uid={username})(cn=\\7busername\\7d))",
		"(uid={username}",
	} {
		_, err := ValidateLoginFilter(filter, "uid")
		require.ErrorIs(t, err, ErrInvalidLoginFilter, filter)
	}
}

func TestIdentityAttributeWhitelist(t *testing.T) {
	require.True(t, IsIdentityAttribute("entryUUID"))
	require.True(t, IsIdentityAttribute("nsUniqueId"))
	require.True(t, IsIdentityAttribute("ipaUniqueID"))
	require.False(t, IsIdentityAttribute("uid"))
	require.False(t, IsIdentityAttribute("objectGUID"))
}
