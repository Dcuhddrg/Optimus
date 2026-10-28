package system_test

import (
	"testing"

	"optimus/server/config"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
)

func TestIssueAndParseJWT(t *testing.T) {
	cfg := testIdentityJWTConfig()
	issuer := systemService.JWTIssuer{Config: cfg}

	token, err := issuer.IssueToken(systemService.TokenSubject{
		ID:        7,
		UUID:      "user-uuid",
		Username:  "admin",
		RoleIDs:   []uint{1, 2},
		RoleCodes: []string{"admin", "ops"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := systemService.ParseJWT(token, cfg)
	require.NoError(t, err)
	require.Equal(t, uint(7), claims.BaseClaims.ID)
	require.Equal(t, "user-uuid", claims.BaseClaims.UUID)
	require.Equal(t, "admin", claims.BaseClaims.Username)
	require.Equal(t, []uint{1, 2}, claims.BaseClaims.RoleIDs)
	require.Equal(t, []string{"admin", "ops"}, claims.BaseClaims.RoleCodes)
	require.Equal(t, "optimus-test", claims.Issuer)
}

func testIdentityJWTConfig() config.JWT {
	return config.JWT{
		SigningKey:  "test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-test",
	}
}
