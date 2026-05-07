package system

import (
	"testing"
	"time"

	"optimus/server/config"

	"github.com/stretchr/testify/require"
)

func TestIssueAndParseJWT(t *testing.T) {
	cfg := testJWTConfig()
	now := time.Now()

	token, err := IssueJWT(TokenSubject{
		ID:        7,
		UUID:      "user-uuid",
		Username:  "admin",
		RoleIDs:   []uint{1, 2},
		RoleCodes: []string{"admin", "ops"},
	}, cfg, now)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ParseJWT(token, cfg)
	require.NoError(t, err)
	require.Equal(t, uint(7), claims.BaseClaims.ID)
	require.Equal(t, "user-uuid", claims.BaseClaims.UUID)
	require.Equal(t, "admin", claims.BaseClaims.Username)
	require.Equal(t, []uint{1, 2}, claims.BaseClaims.RoleIDs)
	require.Equal(t, []string{"admin", "ops"}, claims.BaseClaims.RoleCodes)
	require.Equal(t, "optimus-test", claims.Issuer)
}

func TestParseJWTRejectsExpiredToken(t *testing.T) {
	cfg := testJWTConfig()
	now := time.Now()

	token, err := IssueJWT(TokenSubject{
		ID:       7,
		UUID:     "user-uuid",
		Username: "admin",
	}, cfg, now.Add(-2*time.Hour))
	require.NoError(t, err)

	claims, err := ParseJWT(token, cfg)
	require.Error(t, err)
	require.Nil(t, claims)
}

func testJWTConfig() config.JWT {
	return config.JWT{
		SigningKey:  "test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-test",
	}
}
