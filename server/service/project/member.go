package project

import (
	"context"
	"errors"
	"optimus/server/global"
	projectModel "optimus/server/model/project"
	projectResponse "optimus/server/model/project/response"

	"gorm.io/gorm"
)

func (s *ProjectService) AddMember(ctx context.Context, actorID uint, code string, userID uint, role string) error {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	if !IsValidProjectRole(role) {
		return ErrInvalidProjectRole
	}
	if role == RoleOwner {
		if err := refuseExternalOwner(ctx, userID); err != nil {
			return err
		}
	}
	var users int64
	if err := global.OPTIMUS_DB.WithContext(ctx).
		Table("sys_users").
		Where("id = ?", userID).
		Count(&users).Error; err != nil {
		return err
	}
	if users == 0 {
		return ErrUserNotFound
	}
	m := projectModel.UserProject{
		UserID:      userID,
		ProjectID:   p.ID,
		ProjectRole: role,
	}
	return global.OPTIMUS_DB.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", m.UserID, m.ProjectID).
		Assign(map[string]any{"project_role": m.ProjectRole}).
		FirstOrCreate(&m).Error
}

func (s *ProjectService) UpdateMemberRole(ctx context.Context, actorID uint, code string, userID uint, role string) error {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	if !IsValidProjectRole(role) {
		return ErrInvalidProjectRole
	}
	if role == RoleOwner {
		if err := refuseExternalOwner(ctx, userID); err != nil {
			return err
		}
	}
	existing, err := s.findMember(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	if existing.ProjectRole == RoleOwner && role != RoleOwner {
		owners, err := s.countOwners(ctx, p.ID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}
	if _, err := s.findMember(ctx, userID, p.ID); err != nil {
		return err
	}
	return global.OPTIMUS_DB.WithContext(ctx).
		Model(&projectModel.UserProject{}).
		Where("user_id = ? AND project_id = ?", userID, p.ID).
		Update("project_role", role).Error
}

func (s *ProjectService) RemoveMember(ctx context.Context, actorID uint, code string, userID uint) error {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	existing, err := s.findMember(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	if existing.ProjectRole == RoleOwner {
		owners, err := s.countOwners(ctx, p.ID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}
	res := global.OPTIMUS_DB.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, p.ID).
		Delete(&projectModel.UserProject{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (s *ProjectService) ListMembers(ctx context.Context, actorID uint, code string) ([]projectResponse.ProjectMember, error) {
	p, err := s.findByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if _, err := s.findMember(ctx, actorID, p.ID); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	var rows []projectResponse.ProjectMember
	err = global.OPTIMUS_DB.WithContext(ctx).
		Table("user_projects").
		Select("user_projects.user_id, sys_users.username, sys_users.display_name, user_projects.project_role").
		Joins("JOIN sys_users ON sys_users.id = user_projects.user_id").
		Where("user_projects.project_id = ?", p.ID).
		Scan(&rows).Error
	if rows == nil {
		rows = []projectResponse.ProjectMember{}
	}
	return rows, err
}

func (s *ProjectService) ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	err := global.OPTIMUS_DB.WithContext(ctx).
		Table("user_projects").
		Where("user_id = ?", userID).
		Pluck("project_id", &ids).Error
	if ids == nil {
		ids = []uint{}
	}
	return ids, err
}

func (s *ProjectService) requireOwner(ctx context.Context, actorID, projectID uint) error {
	m, err := s.findMember(ctx, actorID, projectID)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return ErrForbidden
		}
		return err
	}
	if m.ProjectRole != RoleOwner {
		return ErrForbidden
	}
	return nil
}

// refuseExternalOwner rejects making an external account a project owner
// (user management design 3.6). An owner who later becomes external keeps
// ownership until it is handed over; the sync preview points it out.
func refuseExternalOwner(ctx context.Context, userID uint) error {
	var personType string
	err := global.OPTIMUS_DB.WithContext(ctx).Table("sys_users").Select("person_type").Where("id = ?", userID).Scan(&personType).Error
	if err != nil {
		return err
	}
	if personType == "external" {
		return ErrExternalCannotOwn
	}
	return nil
}

func (s *ProjectService) findMember(ctx context.Context, userID, projectID uint) (projectModel.UserProject, error) {
	var m projectModel.UserProject
	err := global.OPTIMUS_DB.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return projectModel.UserProject{}, ErrMemberNotFound
	}
	return m, err
}

func (s *ProjectService) countOwners(ctx context.Context, projectID uint) (int64, error) {
	var n int64
	err := global.OPTIMUS_DB.WithContext(ctx).
		Model(&projectModel.UserProject{}).
		Where("project_id = ? AND project_role = ?", projectID, RoleOwner).
		Count(&n).Error
	return n, err
}
