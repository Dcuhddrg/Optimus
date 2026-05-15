package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	projectModel "optimus/server/model/project"
)

var projectCodeRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

type ProjectService struct {
	store ProjectStore
}

func NewProjectService(store ProjectStore) *ProjectService {
	return &ProjectService{store: store}
}

func (s *ProjectService) Create(ctx context.Context, actorID uint, req CreateProjectRequest) (ProjectView, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 128 {
		return ProjectView{}, ErrInvalidProjectName
	}
	if !projectCodeRegex.MatchString(req.Code) {
		return ProjectView{}, ErrInvalidProjectCode
	}
	p := &projectModel.Project{
		Name:        name,
		Code:        req.Code,
		Description: req.Description,
		Status:      statusActive,
	}
	if err := s.store.Create(ctx, p, actorID); err != nil {
		return ProjectView{}, err
	}
	return toProjectView(*p), nil
}

func (s *ProjectService) List(ctx context.Context, actorID uint) ([]ProjectView, error) {
	ids, err := s.store.ListAccessibleProjectIDs(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []ProjectView{}, nil
	}
	projects, err := s.store.List(ctx, ids)
	if err != nil {
		return nil, err
	}
	views := make([]ProjectView, len(projects))
	for i, p := range projects {
		views[i] = toProjectView(p)
	}
	return views, nil
}

func (s *ProjectService) Get(ctx context.Context, actorID uint, code string) (ProjectView, error) {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return ProjectView{}, err
	}
	if _, err := s.store.FindMember(ctx, actorID, p.ID); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return ProjectView{}, ErrForbidden
		}
		return ProjectView{}, err
	}
	return toProjectView(p), nil
}

func (s *ProjectService) Update(ctx context.Context, actorID uint, code string, req UpdateProjectRequest) (ProjectView, error) {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return ProjectView{}, err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return ProjectView{}, err
	}
	fields := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > 128 {
			return ProjectView{}, ErrInvalidProjectName
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
	updated, err := s.store.Update(ctx, code, fields)
	if err != nil {
		return ProjectView{}, err
	}
	return toProjectView(updated), nil
}

func (s *ProjectService) Delete(ctx context.Context, actorID uint, code string) error {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	count, err := s.store.CountCloudAccounts(ctx, p.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrProjectHasAccounts
	}
	return s.store.Delete(ctx, code)
}

func (s *ProjectService) AddMember(ctx context.Context, actorID uint, code string, userID uint, role string) error {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	if !IsValidProjectRole(role) {
		return ErrInvalidProjectRole
	}
	exists, err := s.store.UserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}
	return s.store.AddOrReplaceMember(ctx, projectModel.UserProject{
		UserID:      userID,
		ProjectID:   p.ID,
		ProjectRole: role,
	})
}

func (s *ProjectService) UpdateMemberRole(ctx context.Context, actorID uint, code string, userID uint, role string) error {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	if !IsValidProjectRole(role) {
		return ErrInvalidProjectRole
	}
	existing, err := s.store.FindMember(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	if existing.ProjectRole == RoleOwner && role != RoleOwner {
		owners, err := s.store.CountOwners(ctx, p.ID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}
	return s.store.UpdateMemberRole(ctx, userID, p.ID, role)
}

func (s *ProjectService) RemoveMember(ctx context.Context, actorID uint, code string, userID uint) error {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if err := s.requireOwner(ctx, actorID, p.ID); err != nil {
		return err
	}
	existing, err := s.store.FindMember(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	if existing.ProjectRole == RoleOwner {
		owners, err := s.store.CountOwners(ctx, p.ID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}
	return s.store.RemoveMember(ctx, userID, p.ID)
}

func (s *ProjectService) ListMembers(ctx context.Context, actorID uint, code string) ([]ProjectMember, error) {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.FindMember(ctx, actorID, p.ID); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	return s.store.ListMembers(ctx, p.ID)
}

func (s *ProjectService) ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error) {
	return s.store.ListAccessibleProjectIDs(ctx, userID)
}

func (s *ProjectService) requireOwner(ctx context.Context, actorID, projectID uint) error {
	m, err := s.store.FindMember(ctx, actorID, projectID)
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

func toProjectView(p projectModel.Project) ProjectView {
	return ProjectView{
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
