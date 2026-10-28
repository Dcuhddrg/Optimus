package system

import (
	"time"

	"optimus/server/config"
)

func IssueJWT(subject TokenSubject, cfg config.JWT, now time.Time) (string, error) {
	return JWTIssuer{Config: cfg, Now: func() time.Time { return now }}.IssueToken(subject)
}
