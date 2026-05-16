package usecase

import (
	"context"
	"errors"
)

func (s *ProjectService) ResolveProjectByCode(ctx context.Context, code string) (uint, bool, error) {
	project, err := s.store.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return project.ID, true, nil
}

func (s *ProjectService) GetMemberRole(ctx context.Context, userID, projectID uint) (string, bool, error) {
	member, err := s.store.FindMember(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return member.ProjectRole, true, nil
}
