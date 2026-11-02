package response

import (
	"time"
)

// AccountView is the API-safe representation (credential masked).
type AccountView struct {
	ID             uint              `json:"id"`
	ProjectID      uint              `json:"projectId"`
	Name           string            `json:"name"`
	Provider       string            `json:"provider"`
	AccountID      string            `json:"accountId"`
	CredentialType string            `json:"credentialType"`
	Credential     string            `json:"credential"`
	Tags           map[string]string `json:"tags"`
	Status         string            `json:"status"`
	LastSyncedAt   *time.Time        `json:"lastSyncedAt"`
}
