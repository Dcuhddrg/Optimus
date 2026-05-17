package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrAccountNotFound  = errors.New("cloud account not found")
	ErrInvalidTagKey    = errors.New("tag key must match ^[a-z][a-z0-9_-]{0,62}$")
	ErrInvalidTagValue  = errors.New("tag value must be 1-128 chars")
	ErrTooManyTags      = errors.New("a cloud account cannot have more than 32 tags")
	ErrInvalidTagFilter = errors.New("invalid tag filter; expected key=value, key=*, or __none__:key")
)

const MaxAccountTags = 32

var tagKeyRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`)

// ValidateTags normalizes and validates a tag map. Returns a cleaned copy.
func ValidateTags(tags map[string]string) (map[string]string, error) {
	if tags == nil {
		return map[string]string{}, nil
	}
	if len(tags) > MaxAccountTags {
		return nil, ErrTooManyTags
	}
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		if !tagKeyRegex.MatchString(k) {
			return nil, ErrInvalidTagKey
		}
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || len(trimmed) > 128 {
			return nil, ErrInvalidTagValue
		}
		out[k] = trimmed
	}
	return out, nil
}

// TagFilter is a single tag filter clause used in list queries.
type TagFilter struct {
	Key    string
	Value  string
	Exists bool
}

// ParseTagFilter parses one tag filter token. Accepted forms:
//   - key=value    => Exists=true, Value=value
//   - key=*        => Exists=true, Value=""
//   - __none__:key => Exists=false
func ParseTagFilter(token string) (TagFilter, error) {
	if strings.HasPrefix(token, "__none__:") {
		key := strings.TrimPrefix(token, "__none__:")
		if !tagKeyRegex.MatchString(key) {
			return TagFilter{}, ErrInvalidTagKey
		}
		return TagFilter{Key: key, Exists: false}, nil
	}
	idx := strings.Index(token, "=")
	if idx <= 0 || idx == len(token)-1 {
		return TagFilter{}, ErrInvalidTagFilter
	}
	key := token[:idx]
	value := token[idx+1:]
	if !tagKeyRegex.MatchString(key) {
		return TagFilter{}, ErrInvalidTagKey
	}
	if value == "*" {
		return TagFilter{Key: key, Value: "", Exists: true}, nil
	}
	if len(value) > 128 {
		return TagFilter{}, ErrInvalidTagValue
	}
	return TagFilter{Key: key, Value: value, Exists: true}, nil
}

// AccountRow is the raw store-level representation (credential is encrypted).
type AccountRow struct {
	ID             uint
	ProjectID      uint
	Name           string
	Provider       string
	AccountID      string
	CredentialType string
	CredentialJSON string
	Tags           map[string]string
	Status         string
	LastSyncedAt   *time.Time
}

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

// AccountStore is the persistence interface for cloud accounts.
type AccountStore interface {
	Create(ctx context.Context, projectID uint, row AccountRow) (AccountRow, error)
	Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (AccountRow, error)
	GetForSync(ctx context.Context, id uint) (AccountRow, error)
	List(ctx context.Context, accessibleProjectIDs []uint, tagFilters []TagFilter) ([]AccountRow, error)
	Update(ctx context.Context, accessibleProjectIDs []uint, id uint, name *string, encryptedCredential *string) (AccountRow, error)
	UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, tags map[string]string) (AccountRow, error)
	Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error
	UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error
}

// CloudAccountService manages cloud accounts and credential encryption.
type CloudAccountService struct {
	store  AccountStore
	encKey []byte
}

func NewCloudAccountService(store AccountStore, encKey []byte) (*CloudAccountService, error) {
	if len(encKey) != 32 {
		return nil, ErrInvalidEncryptionKey
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, encKey)
	return &CloudAccountService{store: store, encKey: keyCopy}, nil
}

func (s *CloudAccountService) Create(ctx context.Context, projectID uint, req CreateAccountRequest) (AccountView, error) {
	credJSON, err := json.Marshal(req.Credential)
	if err != nil {
		return AccountView{}, err
	}
	encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
	if err != nil {
		return AccountView{}, err
	}
	cleanedTags, err := ValidateTags(req.Tags)
	if err != nil {
		return AccountView{}, err
	}
	row := AccountRow{
		ProjectID:      projectID,
		Name:           req.Name,
		Provider:       req.Provider,
		AccountID:      req.AccountID,
		CredentialType: req.CredentialType,
		CredentialJSON: encrypted,
		Tags:           cleanedTags,
	}
	saved, err := s.store.Create(ctx, projectID, row)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(saved), nil
}

// Get returns ErrAccountNotFound if no account exists with the given id.
func (s *CloudAccountService) Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (AccountView, error) {
	row, err := s.store.Get(ctx, accessibleProjectIDs, id)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) List(ctx context.Context, accessibleProjectIDs []uint, tagFilters []TagFilter) ([]AccountView, error) {
	rows, err := s.store.List(ctx, accessibleProjectIDs, tagFilters)
	if err != nil {
		return nil, err
	}
	views := make([]AccountView, len(rows))
	for i, r := range rows {
		views[i] = toAccountView(r)
	}
	return views, nil
}

func (s *CloudAccountService) Update(ctx context.Context, accessibleProjectIDs []uint, id uint, req UpdateAccountRequest) (AccountView, error) {
	var encCredential *string
	if len(req.Credential) > 0 {
		credJSON, err := json.Marshal(req.Credential)
		if err != nil {
			return AccountView{}, err
		}
		enc, err := encryptAESGCM(s.encKey, string(credJSON))
		if err != nil {
			return AccountView{}, err
		}
		encCredential = &enc
	}
	row, err := s.store.Update(ctx, accessibleProjectIDs, id, req.Name, encCredential)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, req UpdateAccountTagsRequest) (AccountView, error) {
	cleaned, err := ValidateTags(req.Tags)
	if err != nil {
		return AccountView{}, err
	}
	row, err := s.store.UpdateTags(ctx, accessibleProjectIDs, id, cleaned)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error {
	return s.store.Delete(ctx, accessibleProjectIDs, id)
}

// UpdateSyncStatus updates last_synced_at and status after a sync attempt.
func (s *CloudAccountService) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	return s.store.UpdateSyncStatus(ctx, id, lastSyncedAt, status)
}

// GetForSync returns a SyncAccount with decrypted credentials for use by AssetRegistryService.
func (s *CloudAccountService) GetForSync(ctx context.Context, id uint) (SyncAccount, error) {
	row, err := s.store.GetForSync(ctx, id)
	if err != nil {
		return SyncAccount{}, err
	}
	plaintext, err := decryptAESGCM(s.encKey, row.CredentialJSON)
	if err != nil {
		return SyncAccount{}, err
	}
	var cred map[string]string
	if err = json.Unmarshal([]byte(plaintext), &cred); err != nil {
		return SyncAccount{}, err
	}
	return SyncAccount{
		ID:             row.ID,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		Credential:     cred,
	}, nil
}

func toAccountView(row AccountRow) AccountView {
	tags := row.Tags
	if tags == nil {
		tags = map[string]string{}
	}
	return AccountView{
		ID:             row.ID,
		ProjectID:      row.ProjectID,
		Name:           row.Name,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		Credential:     "***",
		Tags:           tags,
		Status:         row.Status,
		LastSyncedAt:   row.LastSyncedAt,
	}
}
