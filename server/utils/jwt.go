package utils

import (
	"errors"
	"time"

	"optimus/server/global"

	"github.com/golang-jwt/jwt/v5"
)

// JWT signs and parses tokens with the configured key, as GVA's utils.JWT.
// Token transport, blacklist and refresh are tracked separately in #56.
type JWT struct {
	SigningKey []byte
}

func NewJWT() *JWT {
	return &JWT{SigningKey: []byte(global.OPTIMUS_CONFIG.JWT.SigningKey)}
}

// CreateClaims stamps issuer and lifetime from global JWT config.
func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	cfg := global.OPTIMUS_CONFIG.JWT
	expires := cfg.ExpiresTime
	if expires <= 0 {
		expires = 3600
	}
	now := time.Now()
	return CustomClaims{
		BaseClaims: baseClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expires) * time.Second)),
		},
	}
}

func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	if len(j.SigningKey) == 0 {
		return "", errors.New("jwt signing key is empty")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	if len(j.SigningKey) == 0 {
		return nil, errors.New("jwt signing key is empty")
	}
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected jwt signing method")
		}
		return j.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}
	return claims, nil
}
