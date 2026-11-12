package project

import (
	"context"
	"errors"
	"optimus/server/global"
	projectModel "optimus/server/model/project"
	projectResponse "optimus/server/model/project/response"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if role == RoleOwner {
			if err := refuseExternalOwner(tx, userID); err != nil {
				return err
			}
		}
		var users int64
		if err := tx.Table("sys_users").Where("id = ?", userID).Count(&users).Error; err != nil {
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
		return tx.Where("user_id = ? AND project_id = ?", m.UserID, m.ProjectID).
			Assign(map[string]any{"project_role": m.ProjectRole}).
			FirstOrCreate(&m).Error
	})
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
	return global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if role == RoleOwner {
			if err := refuseExternalOwner(tx, userID); err != nil {
				return err
			}
		}
		return tx.Model(&projectModel.UserProject{}).
			Where("user_id = ? AND project_id = ?", userID, p.ID).
			Update("project_role", role).Error
	})
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

// ownerLockedHook lets tests pause after the user row is locked.
var ownerLockedHook func(userID uint)

// refuseExternalOwner rejects making an external account a project owner
// (user management design 3.6). It locks the user row (SELECT ... FOR
// UPDATE) in tx, which must also write the ownership: a person type change
// by LDAP login or sync updates that row, so it either commits first and the
// owner write is refused, or waits until the ownership is committed, and the
// sync then lists the project for hand-over. An owner who later becomes
// external keeps ownership until it is handed over.
func refuseExternalOwner(tx *gorm.DB, userID uint) error {
	var rows []struct{ PersonType string }
	err := tx.Table("sys_users").Select("person_type").Where("id = ?", userID).Clauses(clause.Locking{Strength: "UPDATE"}).Scan(&rows).Error
	if err != nil {
		return err
	}
	if hook := ownerLockedHook; hook != nil {
		hook(userID)
	}
	if len(rows) > 0 && rows[0].PersonType == "external" {
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
