package persistence

import (
	"context"
	"encoding/json"
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

func (s *GormAccountStore) Create(ctx context.Context, projectID uint, row usecase.AccountRow) (usecase.AccountRow, error) {
	tagsJSON, err := marshalTags(row.Tags)
	if err != nil {
		return usecase.AccountRow{}, err
	}
	m := cloudModel.CloudAccount{
		ProjectID:      projectID,
		Name:           row.Name,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		CredentialJSON: row.CredentialJSON,
		TagsJSON:       tagsJSON,
		Status:         "active",
	}
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m)
}

func (s *GormAccountStore) Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m)
}

func (s *GormAccountStore) GetForSync(ctx context.Context, id uint) (usecase.AccountRow, error) {
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m)
}

func (s *GormAccountStore) List(ctx context.Context, accessibleProjectIDs []uint, filters []usecase.TagFilter) ([]usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return []usecase.AccountRow{}, nil
	}
	tx := s.db.WithContext(ctx).Where("project_id IN ?", accessibleProjectIDs)
	for _, filter := range filters {
		expr, args := tagFilterToSQL(filter)
		tx = tx.Where(expr, args...)
	}
	var models []cloudModel.CloudAccount
	if err := tx.
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	rows := make([]usecase.AccountRow, 0, len(models))
	for _, m := range models {
		row, err := toAccountRow(m)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *GormAccountStore) Update(ctx context.Context, accessibleProjectIDs []uint, id uint, name *string, encryptedCredential *string) (usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&m).Error; err != nil {
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
		if err := s.db.WithContext(ctx).First(&m, m.ID).Error; err != nil {
			return usecase.AccountRow{}, err
		}
	}
	return toAccountRow(m)
}

func (s *GormAccountStore) UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, tags map[string]string) (usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	tagsJSON, err := marshalTags(tags)
	if err != nil {
		return usecase.AccountRow{}, err
	}
	if err := s.db.WithContext(ctx).Model(&m).Update("tags_json", tagsJSON).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	m.TagsJSON = tagsJSON
	return toAccountRow(m)
}

func (s *GormAccountStore) Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error {
	if len(accessibleProjectIDs) == 0 {
		return usecase.ErrAccountNotFound
	}
	res := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		Delete(&cloudModel.CloudAccount{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return usecase.ErrAccountNotFound
	}
	return nil
}

func (s *GormAccountStore) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	updates := map[string]any{"status": status}
	if lastSyncedAt != nil {
		updates["last_synced_at"] = lastSyncedAt
	}
	res := s.db.WithContext(ctx).Model(&cloudModel.CloudAccount{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return usecase.ErrAccountNotFound
	}
	return nil
}

func toAccountRow(m cloudModel.CloudAccount) (usecase.AccountRow, error) {
	tags, err := unmarshalTags(m.TagsJSON)
	if err != nil {
		return usecase.AccountRow{}, err
	}
	return usecase.AccountRow{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Name:           m.Name,
		Provider:       m.Provider,
		AccountID:      m.AccountID,
		CredentialType: m.CredentialType,
		CredentialJSON: m.CredentialJSON,
		Tags:           tags,
		Status:         m.Status,
		LastSyncedAt:   m.LastSyncedAt,
	}, nil
}

func marshalTags(tags map[string]string) (string, error) {
	if tags == nil {
		return "{}", nil
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalTags(s string) (map[string]string, error) {
	if s == "" {
		return map[string]string{}, nil
	}
	var out map[string]string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]string{}
	}
	return out, nil
}

func tagFilterToSQL(f usecase.TagFilter) (string, []any) {
	path := "$." + f.Key
	if !f.Exists {
		return "JSON_EXTRACT(tags_json, ?) IS NULL", []any{path}
	}
	if f.Value == "" {
		return "JSON_EXTRACT(tags_json, ?) IS NOT NULL", []any{path}
	}
	return "JSON_UNQUOTE(JSON_EXTRACT(tags_json, ?)) = ?", []any{path, f.Value}
}
