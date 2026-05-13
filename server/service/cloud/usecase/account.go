package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrAccountNotFound = errors.New("cloud account not found")

// AccountRow is the raw store-level representation (credential is encrypted).
type AccountRow struct {
	ID             uint
	Name           string
	Provider       string
	AccountID      string
	CredentialType string
	CredentialJSON string
	Status         string
	LastSyncedAt   *time.Time
}

// AccountView is the API-safe representation (credential masked).
type AccountView struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	Provider       string     `json:"provider"`
	AccountID      string     `json:"accountId"`
	CredentialType string     `json:"credentialType"`
	Credential     string     `json:"credential"`
	Status         string     `json:"status"`
	LastSyncedAt   *time.Time `json:"lastSyncedAt"`
}

// CreateAccountRequest is the body for POST /cloud/accounts.
type CreateAccountRequest struct {
	Name           string            `json:"name" binding:"required"`
	Provider       string            `json:"provider" binding:"required"`
	AccountID      string            `json:"accountId" binding:"required"`
	CredentialType string            `json:"credentialType" binding:"required"`
	Credential     map[string]string `json:"credential" binding:"required"`
}

// UpdateAccountRequest is the body for PUT /cloud/accounts/:id.
type UpdateAccountRequest struct {
	Name       *string           `json:"name"`
	Credential map[string]string `json:"credential"`
}

// AccountStore is the persistence interface for cloud accounts.
type AccountStore interface {
	Create(ctx context.Context, name, provider, accountID, credentialType, encryptedCredential string) (AccountRow, error)
	Get(ctx context.Context, id uint) (AccountRow, error)
	List(ctx context.Context) ([]AccountRow, error)
	Update(ctx context.Context, id uint, name *string, encryptedCredential *string) (AccountRow, error)
	Delete(ctx context.Context, id uint) error
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

func (s *CloudAccountService) Create(ctx context.Context, req CreateAccountRequest) (AccountView, error) {
	credJSON, err := json.Marshal(req.Credential)
	if err != nil {
		return AccountView{}, err
	}
	encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
	if err != nil {
		return AccountView{}, err
	}
	row, err := s.store.Create(ctx, req.Name, req.Provider, req.AccountID, req.CredentialType, encrypted)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

// Get returns ErrAccountNotFound if no account exists with the given id.
func (s *CloudAccountService) Get(ctx context.Context, id uint) (AccountView, error) {
	row, err := s.store.Get(ctx, id)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) List(ctx context.Context) ([]AccountView, error) {
	rows, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]AccountView, len(rows))
	for i, r := range rows {
		views[i] = toAccountView(r)
	}
	return views, nil
}

func (s *CloudAccountService) Update(ctx context.Context, id uint, req UpdateAccountRequest) (AccountView, error) {
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
	row, err := s.store.Update(ctx, id, req.Name, encCredential)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) Delete(ctx context.Context, id uint) error {
	return s.store.Delete(ctx, id)
}

// UpdateSyncStatus updates last_synced_at and status after a sync attempt.
func (s *CloudAccountService) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	return s.store.UpdateSyncStatus(ctx, id, lastSyncedAt, status)
}

// GetForSync returns a SyncAccount with decrypted credentials for use by AssetRegistryService.
func (s *CloudAccountService) GetForSync(ctx context.Context, id uint) (SyncAccount, error) {
	row, err := s.store.Get(ctx, id)
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
	return AccountView{
		ID:             row.ID,
		Name:           row.Name,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		Credential:     "***",
		Status:         row.Status,
		LastSyncedAt:   row.LastSyncedAt,
	}
}
