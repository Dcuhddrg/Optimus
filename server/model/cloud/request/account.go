package request

// AccountSearch is the query for GET /cloud/accounts.
type AccountSearch struct {
	Tags []string `form:"tag"`
}

// CreateAccountRequest is the body for POST /cloud/accounts.
type CreateAccountRequest struct {
	Name           string            `json:"name" binding:"required"`
	Provider       string            `json:"provider" binding:"required"`
	AccountID      string            `json:"accountId" binding:"required"`
	CredentialType string            `json:"credentialType" binding:"required"`
	Credential     map[string]string `json:"credential" binding:"required"`
	Tags           map[string]string `json:"tags"`
}

// UpdateAccountTagsRequest is the body for PUT /cloud/accounts/:id/tags.
type UpdateAccountTagsRequest struct {
	Tags map[string]string `json:"tags" binding:"required"`
}

// UpdateAccountRequest is the body for PUT /cloud/accounts/:id.
type UpdateAccountRequest struct {
	Name       *string           `json:"name"`
	Credential map[string]string `json:"credential"`
}
