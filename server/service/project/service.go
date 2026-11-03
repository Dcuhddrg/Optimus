package project

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"optimus/server/global"
	projectModel "optimus/server/model/project"
	projectRequest "optimus/server/model/project/request"
	projectResponse "optimus/server/model/project/response"

	"gorm.io/gorm"
)

var projectCodeRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

type ProjectService struct{}

func (s *ProjectService) Create(ctx context.Context, actorID uint, req projectRequest.CreateProjectRequest) (projectResponse.ProjectView, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 128 {
		return projectResponse.ProjectView{}, ErrInvalidProjectName
	}
	if !projectCodeRegex.MatchString(req.Code) {
		return projectResponse.ProjectView{}, ErrInvalidProjectCode
	}
	p := &projectModel.Project{
		Name:        name,
		Code:        req.Code,
		Description: req.Description,
		OwnerID:     actorID,
		Status:      statusActive,
	}
	if err := global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			if isDuplicateKeyError(err) {
				return ErrProjectCodeTaken
			}
			return err
		}
		ownerRow := projectModel.UserProject{
			UserID:      actorID,
			ProjectID:   p.ID,
			ProjectRole: RoleOwner,
		}
		return tx.Create(&ownerRow).Error
	}); err != nil {
		return projectResponse.ProjectView{}, err
	}
	return toProjectView(*p), nil
}

func (s *ProjectService) List(ctx context.Context, actorID uint) ([]projectResponse.ProjectView, error) {
	ids, err := s.ListAccessibleProjectIDs(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []projectResponse.ProjectView{}, nil
	}
	var projects []projectModel.Project
	if err := global.OPTIMUS_DB.WithContext(ctx).
		Where("id IN ?", ids).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, err
	}
	views := make([]projectResponse.ProjectView, len(projects))
	for i, p := range projects {
		views[i] = toProjectView(p)
	}
	return views, nil
}

func (s *ProjectService) Get(ctx context.Context, actorID uint, code string) (projectResponse.ProjectView, error) {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return projectResponse.ProjectView{}, err
	}
	if _, err := s.findMember(ctx, actorID, p.ID); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return projectResponse.ProjectView{}, ErrForbidden
		}
		return projectResponse.ProjectView{}, err
	}
	return toProjectView(p), nil
}

func (s *ProjectService) Update(ctx context.Context, actorID uint, code string, req projectRequest.UpdateProjectRequest) (projectResponse.ProjectView, error) {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return projectResponse.ProjectView{}, err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return projectResponse.ProjectView{}, err
	}
	fields := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > 128 {
			return projectResponse.ProjectView{}, ErrInvalidProjectName
		}
		fields["name"] = name
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if len(fields) == 0 {
		return toProjectView(p), nil
	}
	fields["updated_at"] = time.Now()
	if _, err := s.findByCode(ctx, code); err != nil {
		return projectResponse.ProjectView{}, err
	}
	if err := global.OPTIMUS_DB.WithContext(ctx).
		Model(&projectModel.Project{}).
		Where("code = ?", code).
		Updates(fields).Error; err != nil {
		return projectResponse.ProjectView{}, err
	}
	updated, err := s.findByCode(ctx, code)
	if err != nil {
		return projectResponse.ProjectView{}, err
	}
	return toProjectView(updated), nil
}

func (s *ProjectService) Delete(ctx context.Context, actorID uint, code string) error {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	var count int64
	err = global.OPTIMUS_DB.WithContext(ctx).
		Table("cloud_accounts").
		Where("project_id = ? AND deleted_at IS NULL", p.ID).
		Count(&count).Error
	if err != nil && strings.Contains(err.Error(), "Unknown column 'project_id'") {
		err = nil // SP2 还未跑迁移时 cloud_accounts 无 project_id 列，视为 0
	}
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrProjectHasAccounts
	}
	return global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p projectModel.Project
		if err := tx.Where("code = ?", code).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}
		if err := tx.Where("project_id = ?", p.ID).Delete(&projectModel.UserProject{}).Error; err != nil {
			return err
		}
		return tx.Where("code = ?", code).Delete(&projectModel.Project{}).Error
	})
}

func toProjectView(p projectModel.Project) projectResponse.ProjectView {
	return projectResponse.ProjectView{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		OwnerID:     p.OwnerID,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (s *ProjectService) findByCode(ctx context.Context, code string) (projectModel.Project, error) {
	var p projectModel.Project
	if err := global.OPTIMUS_DB.WithContext(ctx).Where("code = ?", code).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return projectModel.Project{}, ErrProjectNotFound
		}
		return projectModel.Project{}, err
	}
	return p, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "duplicate key value")
}
