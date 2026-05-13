package usecase

import (
	"context"
	"errors"
	"sync"
)

var ErrProviderNotRegistered = errors.New("provider not registered")

// RawAsset is the normalized representation returned by any AssetSyncer.
type RawAsset struct {
	ResourceType string
	ResourceID   string
	Name         string
	Region       string
	Status       string
	ConfigJSON   string
	TagsJSON     string
}

// SyncAccount carries decrypted credentials for a cloud account sync operation.
type SyncAccount struct {
	ID             uint
	Provider       string
	AccountID      string
	CredentialType string
	Credential     map[string]string
}

// AssetSyncer fetches all assets for a given cloud account.
type AssetSyncer interface {
	Sync(ctx context.Context, account SyncAccount) ([]RawAsset, error)
}

// SyncerRegistry maps provider name → AssetSyncer.
type SyncerRegistry struct {
	mu      sync.RWMutex
	syncers map[string]AssetSyncer
}

func NewSyncerRegistry() *SyncerRegistry {
	return &SyncerRegistry{syncers: make(map[string]AssetSyncer)}
}

func (r *SyncerRegistry) Register(provider string, s AssetSyncer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.syncers[provider] = s
}

func (r *SyncerRegistry) Get(provider string) (AssetSyncer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.syncers[provider]
	if !ok {
		return nil, ErrProviderNotRegistered
	}
	return s, nil
}
