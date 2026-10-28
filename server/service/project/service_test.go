package project_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	projectModel "optimus/server/model/project"
	projectService "optimus/server/service/project"

	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	mu       sync.Mutex
	projects map[string]projectModel.Project
	members  map[string]map[uint]projectModel.UserProject // code -> userID -> row
	accounts map[uint]int64                               // projectID -> count
	users    map[uint]bool                                // userID -> exists
	nextID   uint
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		projects: map[string]projectModel.Project{},
		members:  map[string]map[uint]projectModel.UserProject{},
		accounts: map[uint]int64{},
		users:    map[uint]bool{1: true, 2: true, 3: true},
		nextID:   100,
	}
}

func (s *fakeStore) Create(_ context.Context, p *projectModel.Project, ownerID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[p.Code]; ok {
		return projectService.ErrProjectCodeTaken
	}
	p.ID = s.nextID
	s.nextID++
	p.OwnerID = ownerID
	p.Status = "active"
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	s.projects[p.Code] = *p
	s.members[p.Code] = map[uint]projectModel.UserProject{
		ownerID: {UserID: ownerID, ProjectID: p.ID, ProjectRole: projectService.RoleOwner, CreatedAt: time.Now()},
	}
	return nil
}

func (s *fakeStore) FindByCode(_ context.Context, code string) (projectModel.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[code]
	if !ok {
		return projectModel.Project{}, projectService.ErrProjectNotFound
	}
	return p, nil
}

func (s *fakeStore) List(_ context.Context, ids []uint) ([]projectModel.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []projectModel.Project{}
	want := map[uint]bool{}
	for _, id := range ids {
		want[id] = true
	}
	for _, p := range s.projects {
		if want[p.ID] {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *fakeStore) Update(_ context.Context, code string, fields map[string]any) (projectModel.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[code]
	if !ok {
		return projectModel.Project{}, projectService.ErrProjectNotFound
	}
	if v, ok := fields["name"].(string); ok {
		p.Name = v
	}
	if v, ok := fields["description"].(string); ok {
		p.Description = v
	}
	p.UpdatedAt = time.Now()
	s.projects[code] = p
	return p, nil
}

func (s *fakeStore) Delete(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[code]; !ok {
		return projectService.ErrProjectNotFound
	}
	delete(s.projects, code)
	delete(s.members, code)
	return nil
}

func (s *fakeStore) CountCloudAccounts(_ context.Context, projectID uint) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accounts[projectID], nil
}

func (s *fakeStore) AddOrReplaceMember(_ context.Context, m projectModel.UserProject) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == m.ProjectID {
			if s.members[code] == nil {
				s.members[code] = map[uint]projectModel.UserProject{}
			}
			m.CreatedAt = time.Now()
			s.members[code][m.UserID] = m
			return nil
		}
	}
	return projectService.ErrProjectNotFound
}

func (s *fakeStore) UpdateMemberRole(_ context.Context, userID, projectID uint, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			m, ok := s.members[code][userID]
			if !ok {
				return projectService.ErrMemberNotFound
			}
			m.ProjectRole = role
			s.members[code][userID] = m
			return nil
		}
	}
	return projectService.ErrProjectNotFound
}

func (s *fakeStore) RemoveMember(_ context.Context, userID, projectID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			if _, ok := s.members[code][userID]; !ok {
				return projectService.ErrMemberNotFound
			}
			delete(s.members[code], userID)
			return nil
		}
	}
	return projectService.ErrProjectNotFound
}

func (s *fakeStore) ListMembers(_ context.Context, projectID uint) ([]projectModel.ProjectMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			out := []projectModel.ProjectMember{}
			for _, m := range s.members[code] {
				out = append(out, projectModel.ProjectMember{
					UserID:      m.UserID,
					Username:    "user-" + string(rune('A'+int(m.UserID))),
					DisplayName: "",
					ProjectRole: m.ProjectRole,
				})
			}
			return out, nil
		}
	}
	return nil, projectService.ErrProjectNotFound
}

func (s *fakeStore) FindMember(_ context.Context, userID, projectID uint) (projectModel.UserProject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			m, ok := s.members[code][userID]
			if !ok {
				return projectModel.UserProject{}, projectService.ErrMemberNotFound
			}
			return m, nil
		}
	}
	return projectModel.UserProject{}, projectService.ErrProjectNotFound
}

func (s *fakeStore) ListAccessibleProjectIDs(_ context.Context, userID uint) ([]uint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []uint{}
	for code, p := range s.projects {
		if _, ok := s.members[code][userID]; ok {
			out = append(out, p.ID)
		}
	}
	return out, nil
}

func (s *fakeStore) CountOwners(_ context.Context, projectID uint) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			n := int64(0)
			for _, m := range s.members[code] {
				if m.ProjectRole == projectService.RoleOwner {
					n++
				}
			}
			return n, nil
		}
	}
	return 0, nil
}

func (s *fakeStore) UserExists(_ context.Context, userID uint) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.users[userID], nil
}

func TestCreate_AssignsActorAsOwner(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)

	view, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{
		Name: "Prod", Code: "prod", Description: "Production",
	})
	require.NoError(t, err)
	require.Equal(t, "prod", view.Code)
	require.Equal(t, uint(1), view.OwnerID)

	members, err := svc.ListMembers(context.Background(), 1, "prod")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, projectService.RoleOwner, members[0].ProjectRole)
}

func TestCreate_RejectsInvalidCode(t *testing.T) {
	svc := projectService.NewProjectService(newFakeStore())
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "X", Code: "Prod"})
	require.ErrorIs(t, err, projectService.ErrInvalidProjectCode)

	_, err = svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "X", Code: "1bad"})
	require.ErrorIs(t, err, projectService.ErrInvalidProjectCode)
}

func TestCreate_RejectsDuplicateCode(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)

	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "prod"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P2", Code: "prod"})
	require.ErrorIs(t, err, projectService.ErrProjectCodeTaken)
}

func TestList_OnlyReturnsAccessible(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 2, projectModel.CreateProjectRequest{Name: "P2", Code: "p2"})
	require.NoError(t, err)

	views, err := svc.List(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, "p1", views[0].Code)
}

func TestUpdate_OnlyOwnerCanUpdate(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	name := "P1-Renamed"
	_, err = svc.Update(context.Background(), 2, "p1", projectModel.UpdateProjectRequest{Name: &name})
	require.ErrorIs(t, err, projectService.ErrForbidden)

	view, err := svc.Update(context.Background(), 1, "p1", projectModel.UpdateProjectRequest{Name: &name})
	require.NoError(t, err)
	require.Equal(t, "P1-Renamed", view.Name)
}

func TestDelete_RejectsIfCloudAccountsBound(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	view, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	store.accounts[view.ID] = 3

	err = svc.Delete(context.Background(), 1, "p1")
	require.ErrorIs(t, err, projectService.ErrProjectHasAccounts)
}

func TestDelete_OnlyOwnerCanDelete(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.Delete(context.Background(), 2, "p1")
	require.ErrorIs(t, err, projectService.ErrForbidden)

	err = svc.Delete(context.Background(), 1, "p1")
	require.NoError(t, err)
}

func TestAddMember_OnlyOwnerCanAdd(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 2, "p1", 3, projectService.RoleMember)
	require.ErrorIs(t, err, projectService.ErrForbidden)

	err = svc.AddMember(context.Background(), 1, "p1", 2, projectService.RoleMember)
	require.NoError(t, err)
}

func TestAddMember_RejectsInvalidRole(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 1, "p1", 2, "admin")
	require.ErrorIs(t, err, projectService.ErrInvalidProjectRole)
}

func TestAddMember_RejectsUnknownUser(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 1, "p1", 999, projectService.RoleMember)
	require.ErrorIs(t, err, projectService.ErrUserNotFound)
}

func TestRemoveMember_PreventsRemovingLastOwner(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	view, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	_ = view

	err = svc.RemoveMember(context.Background(), 1, "p1", 1)
	require.True(t, errors.Is(err, projectService.ErrCannotRemoveLastOwner))
}

func TestUpdateMemberRole_PreventsDemotingLastOwner(t *testing.T) {
	store := newFakeStore()
	svc := projectService.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, projectModel.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.UpdateMemberRole(context.Background(), 1, "p1", 1, projectService.RoleMember)
	require.ErrorIs(t, err, projectService.ErrCannotRemoveLastOwner)
}
