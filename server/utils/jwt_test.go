package utils

import (
	"testing"

	"optimus/server/config"
	"optimus/server/global"

	"github.com/stretchr/testify/require"
)

func TestCreateAndParseToken(t *testing.T) {
	saved := global.OPTIMUS_CONFIG.JWT
	t.Cleanup(func() { global.OPTIMUS_CONFIG.JWT = saved })
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "test-signing-key", ExpiresTime: 3600, BufferTime: 300, Issuer: "optimus-test"}

	j := NewJWT()
	token, err := j.CreateToken(j.CreateClaims(BaseClaims{
		ID:        7,
		UUID:      "user-uuid",
		Username:  "admin",
		RoleIDs:   []uint{1, 2},
		RoleCodes: []string{"admin", "ops"},
	}))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := NewJWT().ParseToken(token)
	require.NoError(t, err)
	require.Equal(t, uint(7), claims.BaseClaims.ID)
	require.Equal(t, "user-uuid", claims.BaseClaims.UUID)
	require.Equal(t, "admin", claims.BaseClaims.Username)
	require.Equal(t, []uint{1, 2}, claims.BaseClaims.RoleIDs)
	require.Equal(t, []string{"admin", "ops"}, claims.BaseClaims.RoleCodes)
	require.Equal(t, "optimus-test", claims.Issuer)
	require.WithinDuration(t, claims.IssuedAt.Time.Add(3600e9), claims.ExpiresAt.Time, 0)
}

func TestTokenRequiresSigningKey(t *testing.T) {
	saved := global.OPTIMUS_CONFIG.JWT
	t.Cleanup(func() { global.OPTIMUS_CONFIG.JWT = saved })
	global.OPTIMUS_CONFIG.JWT = config.JWT{}
	_, err := NewJWT().CreateToken(CustomClaims{})
	require.Error(t, err)
	_, err = NewJWT().ParseToken("x.y.z")
	require.Error(t, err)
}
