package system

import (
	"errors"
	"time"

	"optimus/server/config"
	"optimus/server/utils"

	"github.com/golang-jwt/jwt/v5"
)

type JWTIssuer struct {
	Config config.JWT
	Now    func() time.Time
}

func (i JWTIssuer) IssueToken(subject TokenSubject) (string, error) {
	cfg := i.Config
	if cfg.SigningKey == "" {
		return "", errors.New("jwt signing key is empty")
	}
	if cfg.ExpiresTime <= 0 {
		cfg.ExpiresTime = 3600
	}

	now := time.Now()
	if i.Now != nil {
		now = i.Now()
	}

	claims := utils.CustomClaims{
		BaseClaims: utils.BaseClaims{
			UUID:      subject.UUID,
			ID:        subject.ID,
			Username:  subject.Username,
			RoleIDs:   subject.RoleIDs,
			RoleCodes: subject.RoleCodes,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.ExpiresTime) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.SigningKey))
}

func ParseJWT(tokenString string, cfg config.JWT) (*utils.CustomClaims, error) {
	if cfg.SigningKey == "" {
		return nil, errors.New("jwt signing key is empty")
	}

	claims := &utils.CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected jwt signing method")
		}
		return []byte(cfg.SigningKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	return claims, nil
}
