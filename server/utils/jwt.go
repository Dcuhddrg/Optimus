package utils

import (
	"errors"
	"time"

	"optimus/server/global"

	"github.com/golang-jwt/jwt/v5"
)

// JWT signs and parses tokens with the configured key, as GVA's utils.JWT.
type JWT struct {
	SigningKey []byte
}

var (
	TokenValid            = errors.New("unknown token error")
	TokenExpired          = errors.New("token is expired")
	TokenNotValidYet      = errors.New("token is not active yet")
	TokenMalformed        = errors.New("token is malformed")
	TokenSignatureInvalid = errors.New("token signature is invalid")
	TokenInvalid          = errors.New("token cannot be handled")
)

func NewJWT() *JWT {
	return &JWT{SigningKey: []byte(global.OPTIMUS_CONFIG.JWT.SigningKey)}
}

// CreateClaims stamps issuer, lifetime and refresh buffer from global JWT config.
func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	bf, _ := ParseDuration(global.OPTIMUS_CONFIG.JWT.BufferTime)
	ep, _ := ParseDuration(global.OPTIMUS_CONFIG.JWT.ExpiresTime)
	now := time.Now()
	return CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(bf / time.Second),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    global.OPTIMUS_CONFIG.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-1000)),
			ExpiresAt: jwt.NewNumericDate(now.Add(ep)),
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

// CreateTokenByOldToken issues a replacement token, merging concurrent
// refreshes of the same old token, as GVA.
func (j *JWT) CreateTokenByOldToken(oldToken string, claims CustomClaims) (string, error) {
	v, err, _ := global.OPTIMUS_Concurrency_Control.Do("JWT:"+oldToken, func() (any, error) {
		return j.CreateToken(claims)
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	if len(j.SigningKey) == 0 {
		return nil, TokenInvalid
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, TokenInvalid
		}
		return j.SigningKey, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, TokenExpired
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, TokenMalformed
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, TokenSignatureInvalid
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, TokenNotValidYet
		default:
			return nil, TokenInvalid
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, TokenValid
}
