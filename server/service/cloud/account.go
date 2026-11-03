package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"optimus/server/global"
	cloudModel "optimus/server/model/cloud"
	cloudRequest "optimus/server/model/cloud/request"
	cloudResponse "optimus/server/model/cloud/response"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
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

// CloudAccountService manages cloud accounts and credential encryption.
type CloudAccountService struct {
	encKey []byte
}

func NewCloudAccountService(encKey []byte) (*CloudAccountService, error) {
	if len(encKey) != 32 {
		return nil, ErrInvalidEncryptionKey
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, encKey)
	return &CloudAccountService{encKey: keyCopy}, nil
}

func (s *CloudAccountService) Create(ctx context.Context, projectID uint, req cloudRequest.CreateAccountRequest) (cloudResponse.AccountView, error) {
	credJSON, err := json.Marshal(req.Credential)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	cleanedTags, err := ValidateTags(req.Tags)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	tagsJSON, err := marshalTags(cleanedTags)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	m := cloudModel.CloudAccount{
		ProjectID:      projectID,
		Name:           req.Name,
		Provider:       req.Provider,
		AccountID:      req.AccountID,
		CredentialType: req.CredentialType,
		CredentialJSON: encrypted,
		TagsJSON:       tagsJSON,
		Status:         "active",
	}
	if err := global.OPTIMUS_DB.WithContext(ctx).Create(&m).Error; err != nil {
		return cloudResponse.AccountView{}, err
	}
	return toAccountView(m)
}

// Get returns ErrAccountNotFound if no account exists with the given id.
func (s *CloudAccountService) Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (cloudResponse.AccountView, error) {
	row, err := findAccount(ctx, accessibleProjectIDs, id)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	return toAccountView(row)
}

func (s *CloudAccountService) List(ctx context.Context, accessibleProjectIDs []uint, tagFilters []TagFilter) ([]cloudResponse.AccountView, error) {
	if len(accessibleProjectIDs) == 0 {
		return []cloudResponse.AccountView{}, nil
	}
	tx := global.OPTIMUS_DB.WithContext(ctx).Where("project_id IN ?", accessibleProjectIDs)
	for _, filter := range tagFilters {
		expr, args := tagFilterToSQL(filter)
		tx = tx.Where(expr, args...)
	}
	var models []cloudModel.CloudAccount
	if err := tx.
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	rows := make([]cloudResponse.AccountView, 0, len(models))
	for _, m := range models {
		row, err := toAccountView(m)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *CloudAccountService) Update(ctx context.Context, accessibleProjectIDs []uint, id uint, req cloudRequest.UpdateAccountRequest) (cloudResponse.AccountView, error) {
	var encCredential *string
	if len(req.Credential) > 0 {
		credJSON, err := json.Marshal(req.Credential)
		if err != nil {
			return cloudResponse.AccountView{}, err
		}
		enc, err := encryptAESGCM(s.encKey, string(credJSON))
		if err != nil {
			return cloudResponse.AccountView{}, err
		}
		encCredential = &enc
	}
	m, err := findAccount(ctx, accessibleProjectIDs, id)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if encCredential != nil {
		updates["credential_json"] = *encCredential
	}
	if len(updates) > 0 {
		if err := global.OPTIMUS_DB.WithContext(ctx).Model(&m).Updates(updates).Error; err != nil {
			return cloudResponse.AccountView{}, err
		}
		if err := global.OPTIMUS_DB.WithContext(ctx).First(&m, m.ID).Error; err != nil {
			return cloudResponse.AccountView{}, err
		}
	}
	return toAccountView(m)
}

func (s *CloudAccountService) UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, req cloudRequest.UpdateAccountTagsRequest) (cloudResponse.AccountView, error) {
	cleaned, err := ValidateTags(req.Tags)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	m, err := findAccount(ctx, accessibleProjectIDs, id)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	tagsJSON, err := marshalTags(cleaned)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	if err := global.OPTIMUS_DB.WithContext(ctx).Model(&m).Update("tags_json", tagsJSON).Error; err != nil {
		return cloudResponse.AccountView{}, err
	}
	m.TagsJSON = tagsJSON
	return toAccountView(m)
}

func (s *CloudAccountService) Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error {
	if len(accessibleProjectIDs) == 0 {
		return ErrAccountNotFound
	}
	res := global.OPTIMUS_DB.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		Delete(&cloudModel.CloudAccount{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrAccountNotFound
	}
	return nil
}

// UpdateSyncStatus updates last_synced_at and status after a sync attempt.
func (s *CloudAccountService) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	updates := map[string]any{"status": status}
	if lastSyncedAt != nil {
		updates["last_synced_at"] = lastSyncedAt
	}
	res := global.OPTIMUS_DB.WithContext(ctx).Model(&cloudModel.CloudAccount{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrAccountNotFound
	}
	return nil
}

// GetForSync returns a SyncAccount with decrypted credentials for use by AssetRegistryService.
func (s *CloudAccountService) GetForSync(ctx context.Context, id uint) (SyncAccount, error) {
	var m cloudModel.CloudAccount
	if err := global.OPTIMUS_DB.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SyncAccount{}, ErrAccountNotFound
		}
		return SyncAccount{}, err
	}
	if _, err := unmarshalTags(m.TagsJSON); err != nil {
		return SyncAccount{}, err
	}
	plaintext, err := decryptAESGCM(s.encKey, m.CredentialJSON)
	if err != nil {
		return SyncAccount{}, err
	}
	var cred map[string]string
	if err = json.Unmarshal([]byte(plaintext), &cred); err != nil {
		return SyncAccount{}, err
	}
	return SyncAccount{
		ID:             m.ID,
		Provider:       m.Provider,
		AccountID:      m.AccountID,
		CredentialType: m.CredentialType,
		Credential:     cred,
	}, nil
}

func toAccountView(m cloudModel.CloudAccount) (cloudResponse.AccountView, error) {
	tags, err := unmarshalTags(m.TagsJSON)
	if err != nil {
		return cloudResponse.AccountView{}, err
	}
	return cloudResponse.AccountView{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Name:           m.Name,
		Provider:       m.Provider,
		AccountID:      m.AccountID,
		CredentialType: m.CredentialType,
		Credential:     "***",
		Tags:           tags,
		Status:         m.Status,
		LastSyncedAt:   m.LastSyncedAt,
	}, nil
}

func findAccount(ctx context.Context, accessibleProjectIDs []uint, id uint) (cloudModel.CloudAccount, error) {
	if len(accessibleProjectIDs) == 0 {
		return cloudModel.CloudAccount{}, ErrAccountNotFound
	}
	var m cloudModel.CloudAccount
	if err := global.OPTIMUS_DB.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cloudModel.CloudAccount{}, ErrAccountNotFound
		}
		return cloudModel.CloudAccount{}, err
	}
	return m, nil
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

func tagFilterToSQL(f TagFilter) (string, []any) {
	path := "$." + f.Key
	if !f.Exists {
		return "JSON_EXTRACT(tags_json, ?) IS NULL", []any{path}
	}
	if f.Value == "" {
		return "JSON_EXTRACT(tags_json, ?) IS NOT NULL", []any{path}
	}
	return "JSON_UNQUOTE(JSON_EXTRACT(tags_json, ?)) = ?", []any{path, f.Value}
}
