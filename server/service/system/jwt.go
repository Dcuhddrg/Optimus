package system

import (
	"time"

	"optimus/server/config"
	"optimus/server/service/system/infra/identity"
	"optimus/server/service/system/usecase"
	"optimus/server/utils"
)

// TokenSubject is a temporary alias so callers in the root package still compile
// while the real auth logic is migrated to usecase in a later task.
type TokenSubject = usecase.TokenSubject

func IssueJWT(subject usecase.TokenSubject, cfg config.JWT, now time.Time) (string, error) {
	return identity.JWTIssuer{Config: cfg, Now: func() time.Time { return now }}.IssueToken(subject)
}

func ParseJWT(tokenString string, cfg config.JWT) (*utils.CustomClaims, error) {
	return identity.ParseJWT(tokenString, cfg)
}
