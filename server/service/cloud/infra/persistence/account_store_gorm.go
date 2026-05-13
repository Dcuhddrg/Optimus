package persistence

import (
	"context"
	"errors"
	"time"

	cloudModel "optimus/server/model/cloud"
	"optimus/server/service/cloud/usecase"

	"gorm.io/gorm"
)

type GormAccountStore struct {
	db *gorm.DB
}

func NewGormAccountStore(db *gorm.DB) *GormAccountStore {
	return &GormAccountStore{db: db}
}

func (s *GormAccountStore) Create(ctx context.Context, name, provider, accountID, credentialType, encryptedCredential string) (usecase.AccountRow, error) {
	m := cloudModel.CloudAccount{
		Name:           name,
		Provider:       provider,
		AccountID:      accountID,
		CredentialType: credentialType,
		CredentialJSON: encryptedCredential,
		Status:         "active",
	}
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m), nil
}

func (s *GormAccountStore) Get(ctx context.Context, id uint) (usecase.AccountRow, error) {
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m), nil
}

func (s *GormAccountStore) List(ctx context.Context) ([]usecase.AccountRow, error) {
	var models []cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	rows := make([]usecase.AccountRow, len(models))
	for i, m := range models {
		rows[i] = toAccountRow(m)
	}
	return rows, nil
}

func (s *GormAccountStore) Update(ctx context.Context, id uint, name *string, encryptedCredential *string) (usecase.AccountRow, error) {
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
	}
	if encryptedCredential != nil {
		updates["credential_json"] = *encryptedCredential
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&m).Updates(updates).Error; err != nil {
			return usecase.AccountRow{}, err
		}
	}
	return toAccountRow(m), nil
}

func (s *GormAccountStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&cloudModel.CloudAccount{}, id).Error
}

func (s *GormAccountStore) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	updates := map[string]any{"status": status}
	if lastSyncedAt != nil {
		updates["last_synced_at"] = lastSyncedAt
	}
	return s.db.WithContext(ctx).Model(&cloudModel.CloudAccount{}).Where("id = ?", id).Updates(updates).Error
}

func toAccountRow(m cloudModel.CloudAccount) usecase.AccountRow {
	return usecase.AccountRow{
		ID:             m.ID,
		Name:           m.Name,
		Provider:       m.Provider,
		AccountID:      m.AccountID,
		CredentialType: m.CredentialType,
		CredentialJSON: m.CredentialJSON,
		Status:         m.Status,
		LastSyncedAt:   m.LastSyncedAt,
	}
}
