package persistence

import (
	"context"
	"errors"
	"strings"

	projectModel "optimus/server/model/project"
	"optimus/server/service/project/usecase"

	"gorm.io/gorm"
)

type GormProjectStore struct {
	db *gorm.DB
}

func NewGormProjectStore(db *gorm.DB) *GormProjectStore {
	return &GormProjectStore{db: db}
}

func (s *GormProjectStore) Create(ctx context.Context, p *projectModel.Project, ownerID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		p.OwnerID = ownerID
		if p.Status == "" {
			p.Status = "active"
		}
		if err := tx.Create(p).Error; err != nil {
			if isDuplicateKeyError(err) {
				return usecase.ErrProjectCodeTaken
			}
			return err
		}
		ownerRow := projectModel.UserProject{
			UserID:      ownerID,
			ProjectID:   p.ID,
			ProjectRole: usecase.RoleOwner,
		}
		return tx.Create(&ownerRow).Error
	})
}

func (s *GormProjectStore) FindByCode(ctx context.Context, code string) (projectModel.Project, error) {
	var p projectModel.Project
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return projectModel.Project{}, usecase.ErrProjectNotFound
		}
		return projectModel.Project{}, err
	}
	return p, nil
}

func (s *GormProjectStore) List(ctx context.Context, accessibleIDs []uint) ([]projectModel.Project, error) {
	if len(accessibleIDs) == 0 {
		return []projectModel.Project{}, nil
	}
	var rows []projectModel.Project
	err := s.db.WithContext(ctx).
		Where("id IN ?", accessibleIDs).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (s *GormProjectStore) Update(ctx context.Context, code string, fields map[string]any) (projectModel.Project, error) {
	var existing projectModel.Project
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return projectModel.Project{}, usecase.ErrProjectNotFound
		}
		return projectModel.Project{}, err
	}
	if err := s.db.WithContext(ctx).
		Model(&projectModel.Project{}).
		Where("code = ?", code).
		Updates(fields).Error; err != nil {
		return projectModel.Project{}, err
	}
	return s.FindByCode(ctx, code)
}

func (s *GormProjectStore) Delete(ctx context.Context, code string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p projectModel.Project
		if err := tx.Where("code = ?", code).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return usecase.ErrProjectNotFound
			}
			return err
		}
		if err := tx.Where("project_id = ?", p.ID).Delete(&projectModel.UserProject{}).Error; err != nil {
			return err
		}
		return tx.Where("code = ?", code).Delete(&projectModel.Project{}).Error
	})
}

func (s *GormProjectStore) CountCloudAccounts(ctx context.Context, projectID uint) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).
		Table("cloud_accounts").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Count(&n).Error
	if err != nil && strings.Contains(err.Error(), "Unknown column 'project_id'") {
		return 0, nil // SP2 还未跑迁移时 cloud_accounts 无 project_id 列，视为 0
	}
	return n, err
}

func (s *GormProjectStore) AddOrReplaceMember(ctx context.Context, m projectModel.UserProject) error {
	return s.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", m.UserID, m.ProjectID).
		Assign(map[string]any{"project_role": m.ProjectRole}).
		FirstOrCreate(&m).Error
}

func (s *GormProjectStore) UpdateMemberRole(ctx context.Context, userID, projectID uint, role string) error {
	var existing projectModel.UserProject
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.ErrMemberNotFound
		}
		return err
	}
	return s.db.WithContext(ctx).
		Model(&projectModel.UserProject{}).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Update("project_role", role).Error
}

func (s *GormProjectStore) RemoveMember(ctx context.Context, userID, projectID uint) error {
	res := s.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Delete(&projectModel.UserProject{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return usecase.ErrMemberNotFound
	}
	return nil
}

func (s *GormProjectStore) ListMembers(ctx context.Context, projectID uint) ([]usecase.ProjectMember, error) {
	var rows []usecase.ProjectMember
	err := s.db.WithContext(ctx).
		Table("user_projects").
		Select("user_projects.user_id, sys_users.username, sys_users.display_name, user_projects.project_role").
		Joins("JOIN sys_users ON sys_users.id = user_projects.user_id").
		Where("user_projects.project_id = ?", projectID).
		Scan(&rows).Error
	if rows == nil {
		rows = []usecase.ProjectMember{}
	}
	return rows, err
}

func (s *GormProjectStore) FindMember(ctx context.Context, userID, projectID uint) (projectModel.UserProject, error) {
	var m projectModel.UserProject
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return projectModel.UserProject{}, usecase.ErrMemberNotFound
	}
	return m, err
}

func (s *GormProjectStore) ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	err := s.db.WithContext(ctx).
		Table("user_projects").
		Where("user_id = ?", userID).
		Pluck("project_id", &ids).Error
	if ids == nil {
		ids = []uint{}
	}
	return ids, err
}

func (s *GormProjectStore) CountOwners(ctx context.Context, projectID uint) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).
		Model(&projectModel.UserProject{}).
		Where("project_id = ? AND project_role = ?", projectID, usecase.RoleOwner).
		Count(&n).Error
	return n, err
}

func (s *GormProjectStore) UserExists(ctx context.Context, userID uint) (bool, error) {
	var n int64
	err := s.db.WithContext(ctx).
		Table("sys_users").
		Where("id = ?", userID).
		Count(&n).Error
	return n > 0, err
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "duplicate key value")
}
