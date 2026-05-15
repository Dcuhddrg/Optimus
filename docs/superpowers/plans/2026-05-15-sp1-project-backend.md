# SP1 Project Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地"项目"实体与成员管理模块，作为多项目隔离的基础设施（无侵入既有 cloud 模块）。

**Architecture:** 复用既有 `system` 模块的 clean-architecture 分层：`model/project` (GORM) → `service/project/usecase` (业务) → `service/project/infra/persistence` (GORM 实现) → `api/v1/project` (HTTP) → `router/project` (路由) → `initialize/project` (装配)。鉴权：actor 必须是项目 owner 才能写操作；成员关系存 `user_projects` 表，三种 role: owner/member/viewer。

**Tech Stack:** Go 1.23, GORM v2, Gin, testify; MySQL（测试需 `OPTIMUS_TEST_MYSQL_DSN` 环境变量）。

**Spec reference:** `docs/superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md` §SP1 + §Data Model + §API Surface。

---

## File Structure

**新增：**
- `server/model/project/project.go` — Project + UserProject GORM models
- `server/service/project/project.go` — AutoMigrateProject + SeedProjectData + DefaultProjectPermissions
- `server/service/project/usecase/types.go` — sentinel errors, role 常量, Store 接口, Request/View 类型
- `server/service/project/usecase/project.go` — ProjectService（CRUD + Member 方法）
- `server/service/project/usecase/project_test.go` — usecase 单元测试（fake store）
- `server/service/project/infra/persistence/project_store_gorm.go` — GORM 实现
- `server/service/project/infra/persistence/project_store_gorm_test.go` — 持久化集成测试
- `server/service/project/infra/persistence/testdb_test.go` — 测试 DB 启动辅助
- `server/api/v1/project/handlers.go` — ProjectHandlers + Deps
- `server/api/v1/project/project.go` — Project 5 个 HTTP 方法
- `server/api/v1/project/member.go` — Member 4 个 HTTP 方法
- `server/router/project/enter.go` — InitProjectRouter
- `server/initialize/project/module.go` — NewProjectModule

**修改：**
- `server/router/enter.go` — 新增 `Project project.RouterGroup` 字段
- `server/initialize/router.go` — 新增 `var newProjectModule = projectInit.NewProjectModule` + 路由注册
- `server/initialize/startup.go` — 调用 `AutoMigrateProject` + `SeedProjectData`

---

### Task 1: GORM 模型 + AutoMigrate

**Files:**
- Create: `server/model/project/project.go`
- Create: `server/service/project/project.go`

- [ ] **Step 1.1: 写失败的 model 编译测试**

文件: `server/model/project/project.go`

```go
package project

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Code        string         `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Description string         `gorm:"size:512" json:"description"`
	OwnerID     uint           `gorm:"not null;index" json:"ownerId"`
	Status      string         `gorm:"size:16;not null;default:'active'" json:"status"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (Project) TableName() string {
	return "projects"
}

type UserProject struct {
	UserID      uint      `gorm:"primaryKey;autoIncrement:false" json:"userId"`
	ProjectID   uint      `gorm:"primaryKey;autoIncrement:false" json:"projectId"`
	ProjectRole string    `gorm:"size:32;not null" json:"projectRole"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (UserProject) TableName() string {
	return "user_projects"
}
```

- [ ] **Step 1.2: 写 AutoMigrateProject + SeedProjectData skeleton**

文件: `server/service/project/project.go`

```go
package project

import (
	"errors"

	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

func AutoMigrateProject(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	return db.AutoMigrate(
		&projectModel.Project{},
		&projectModel.UserProject{},
	)
}

func DefaultProjectPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "project.menu", Name: "Projects", Type: "menu", Path: "/projects", Sort: 30},
		{Code: "project.view", Name: "View projects", Type: "api", Method: "GET", Path: "/api/v1/projects", Sort: 700},
		{Code: "project.detail", Name: "View project detail", Type: "api", Method: "GET", Path: "/api/v1/projects/:code", Sort: 701},
		{Code: "project.create", Name: "Create project", Type: "api", Method: "POST", Path: "/api/v1/projects", Sort: 702},
		{Code: "project.update", Name: "Update project", Type: "api", Method: "PUT", Path: "/api/v1/projects/:code", Sort: 703},
		{Code: "project.delete", Name: "Delete project", Type: "api", Method: "DELETE", Path: "/api/v1/projects/:code", Sort: 704},
		{Code: "project.members.view", Name: "View project members", Type: "api", Method: "GET", Path: "/api/v1/projects/:code/members", Sort: 710},
		{Code: "project.members.add", Name: "Add project member", Type: "api", Method: "POST", Path: "/api/v1/projects/:code/members", Sort: 711},
		{Code: "project.members.update", Name: "Update project member role", Type: "api", Method: "PATCH", Path: "/api/v1/projects/:code/members/:userId", Sort: 712},
		{Code: "project.members.remove", Name: "Remove project member", Type: "api", Method: "DELETE", Path: "/api/v1/projects/:code/members/:userId", Sort: 713},
	}
}

func SeedProjectData(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	for _, perm := range DefaultProjectPermissions() {
		p := perm
		if err := db.Where("code = ?", p.Code).FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}
	return seedAdminProjectPermissions(db)
}

func seedAdminProjectPermissions(db *gorm.DB) error {
	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	for _, perm := range DefaultProjectPermissions() {
		var p systemModel.SystemPermission
		if err := db.Where("code = ?", perm.Code).First(&p).Error; err != nil {
			return err
		}
		rp := systemModel.SystemRolePermission{RoleID: adminRole.ID, PermissionID: p.ID}
		if err := db.Where("role_id = ? AND permission_id = ?", adminRole.ID, p.ID).FirstOrCreate(&rp).Error; err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 1.3: 验证编译通过**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean exit, no errors.

- [ ] **Step 1.4: 提交**

```bash
cd /home/lilin/trae-project/Optimus
git add server/model/project/ server/service/project/project.go
git commit -m "$(cat <<'EOF'
新增 Project 模块 GORM 模型与迁移种子骨架

Constraint: 项目实体作为多项目隔离的基础，无侵入既有 cloud 模块
Confidence: high
Scope-risk: narrow
Tested: go build ./...
EOF
)"
```

---

### Task 2: Usecase 类型与 Store 接口

**Files:**
- Create: `server/service/project/usecase/types.go`

- [ ] **Step 2.1: 写类型契约**

文件: `server/service/project/usecase/types.go`

```go
package usecase

import (
	"context"
	"errors"
	"time"

	projectModel "optimus/server/model/project"
)

var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrProjectCodeTaken      = errors.New("project code already taken")
	ErrInvalidProjectCode    = errors.New("project code must match ^[a-z][a-z0-9-]{1,62}$")
	ErrInvalidProjectName    = errors.New("project name must be 1-128 chars")
	ErrProjectHasAccounts    = errors.New("project has bound cloud accounts; unbind first")
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidProjectRole    = errors.New("project role must be owner/member/viewer")
	ErrMemberNotFound        = errors.New("member not found")
	ErrCannotRemoveLastOwner = errors.New("cannot remove the last project owner")
	ErrUserNotFound          = errors.New("target user does not exist")
)

const (
	RoleOwner  = "owner"
	RoleMember = "member"
	RoleViewer = "viewer"

	statusActive = "active"
)

func IsValidProjectRole(role string) bool {
	return role == RoleOwner || role == RoleMember || role == RoleViewer
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AddMemberRequest struct {
	UserID      uint   `json:"userId"`
	ProjectRole string `json:"projectRole"`
}

type UpdateMemberRoleRequest struct {
	ProjectRole string `json:"projectRole"`
}

type ProjectView struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"ownerId"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProjectMember struct {
	UserID      uint   `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	ProjectRole string `json:"projectRole"`
}

type ProjectStore interface {
	Create(ctx context.Context, p *projectModel.Project, ownerID uint) error
	FindByCode(ctx context.Context, code string) (projectModel.Project, error)
	List(ctx context.Context, accessibleIDs []uint) ([]projectModel.Project, error)
	Update(ctx context.Context, code string, fields map[string]any) (projectModel.Project, error)
	Delete(ctx context.Context, code string) error
	CountCloudAccounts(ctx context.Context, projectID uint) (int64, error)

	AddOrReplaceMember(ctx context.Context, m projectModel.UserProject) error
	UpdateMemberRole(ctx context.Context, userID, projectID uint, role string) error
	RemoveMember(ctx context.Context, userID, projectID uint) error
	ListMembers(ctx context.Context, projectID uint) ([]ProjectMember, error)
	FindMember(ctx context.Context, userID, projectID uint) (projectModel.UserProject, error)
	ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error)
	CountOwners(ctx context.Context, projectID uint) (int64, error)
	UserExists(ctx context.Context, userID uint) (bool, error)
}
```

- [ ] **Step 2.2: 验证编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean exit.

- [ ] **Step 2.3: 提交**

```bash
git add server/service/project/usecase/types.go
git commit -m "$(cat <<'EOF'
定义 Project 模块 usecase 类型契约与 Store 接口

Confidence: high
Scope-risk: narrow
Tested: go build ./...
EOF
)"
```

---

### Task 3: ProjectService CRUD + 单元测试

**Files:**
- Create: `server/service/project/usecase/project.go`
- Create: `server/service/project/usecase/project_test.go`

- [ ] **Step 3.1: 写失败的 CRUD 单元测试**

文件: `server/service/project/usecase/project_test.go`

```go
package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	projectModel "optimus/server/model/project"
	"optimus/server/service/project/usecase"

	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	mu        sync.Mutex
	projects  map[string]projectModel.Project
	members   map[string]map[uint]projectModel.UserProject // code -> userID -> row
	accounts  map[uint]int64                                // projectID -> count
	users     map[uint]bool                                 // userID -> exists
	nextID    uint
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
		return usecase.ErrProjectCodeTaken
	}
	p.ID = s.nextID
	s.nextID++
	p.OwnerID = ownerID
	p.Status = "active"
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	s.projects[p.Code] = *p
	s.members[p.Code] = map[uint]projectModel.UserProject{
		ownerID: {UserID: ownerID, ProjectID: p.ID, ProjectRole: usecase.RoleOwner, CreatedAt: time.Now()},
	}
	return nil
}

func (s *fakeStore) FindByCode(_ context.Context, code string) (projectModel.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[code]
	if !ok {
		return projectModel.Project{}, usecase.ErrProjectNotFound
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
		return projectModel.Project{}, usecase.ErrProjectNotFound
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
		return usecase.ErrProjectNotFound
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
	return usecase.ErrProjectNotFound
}

func (s *fakeStore) UpdateMemberRole(_ context.Context, userID, projectID uint, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			m, ok := s.members[code][userID]
			if !ok {
				return usecase.ErrMemberNotFound
			}
			m.ProjectRole = role
			s.members[code][userID] = m
			return nil
		}
	}
	return usecase.ErrProjectNotFound
}

func (s *fakeStore) RemoveMember(_ context.Context, userID, projectID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			if _, ok := s.members[code][userID]; !ok {
				return usecase.ErrMemberNotFound
			}
			delete(s.members[code], userID)
			return nil
		}
	}
	return usecase.ErrProjectNotFound
}

func (s *fakeStore) ListMembers(_ context.Context, projectID uint) ([]usecase.ProjectMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			out := []usecase.ProjectMember{}
			for _, m := range s.members[code] {
				out = append(out, usecase.ProjectMember{
					UserID:      m.UserID,
					Username:    "user-" + string(rune('A'+int(m.UserID))),
					DisplayName: "",
					ProjectRole: m.ProjectRole,
				})
			}
			return out, nil
		}
	}
	return nil, usecase.ErrProjectNotFound
}

func (s *fakeStore) FindMember(_ context.Context, userID, projectID uint) (projectModel.UserProject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, p := range s.projects {
		if p.ID == projectID {
			m, ok := s.members[code][userID]
			if !ok {
				return projectModel.UserProject{}, usecase.ErrMemberNotFound
			}
			return m, nil
		}
	}
	return projectModel.UserProject{}, usecase.ErrProjectNotFound
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
				if m.ProjectRole == usecase.RoleOwner {
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
	svc := usecase.NewProjectService(store)

	view, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{
		Name: "Prod", Code: "prod", Description: "Production",
	})
	require.NoError(t, err)
	require.Equal(t, "prod", view.Code)
	require.Equal(t, uint(1), view.OwnerID)

	members, err := svc.ListMembers(context.Background(), 1, "prod")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, usecase.RoleOwner, members[0].ProjectRole)
}

func TestCreate_RejectsInvalidCode(t *testing.T) {
	svc := usecase.NewProjectService(newFakeStore())
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "X", Code: "Prod"})
	require.ErrorIs(t, err, usecase.ErrInvalidProjectCode)

	_, err = svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "X", Code: "1bad"})
	require.ErrorIs(t, err, usecase.ErrInvalidProjectCode)
}

func TestCreate_RejectsDuplicateCode(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)

	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "prod"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P2", Code: "prod"})
	require.ErrorIs(t, err, usecase.ErrProjectCodeTaken)
}

func TestList_OnlyReturnsAccessible(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), 2, usecase.CreateProjectRequest{Name: "P2", Code: "p2"})
	require.NoError(t, err)

	views, err := svc.List(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, "p1", views[0].Code)
}

func TestUpdate_OnlyOwnerCanUpdate(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	name := "P1-Renamed"
	_, err = svc.Update(context.Background(), 2, "p1", usecase.UpdateProjectRequest{Name: &name})
	require.ErrorIs(t, err, usecase.ErrForbidden)

	view, err := svc.Update(context.Background(), 1, "p1", usecase.UpdateProjectRequest{Name: &name})
	require.NoError(t, err)
	require.Equal(t, "P1-Renamed", view.Name)
}

func TestDelete_RejectsIfCloudAccountsBound(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	view, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	store.accounts[view.ID] = 3

	err = svc.Delete(context.Background(), 1, "p1")
	require.ErrorIs(t, err, usecase.ErrProjectHasAccounts)
}

func TestDelete_OnlyOwnerCanDelete(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.Delete(context.Background(), 2, "p1")
	require.ErrorIs(t, err, usecase.ErrForbidden)

	err = svc.Delete(context.Background(), 1, "p1")
	require.NoError(t, err)
}

func TestAddMember_OnlyOwnerCanAdd(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 2, "p1", 3, usecase.RoleMember)
	require.ErrorIs(t, err, usecase.ErrForbidden)

	err = svc.AddMember(context.Background(), 1, "p1", 2, usecase.RoleMember)
	require.NoError(t, err)
}

func TestAddMember_RejectsInvalidRole(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 1, "p1", 2, "admin")
	require.ErrorIs(t, err, usecase.ErrInvalidProjectRole)
}

func TestAddMember_RejectsUnknownUser(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.AddMember(context.Background(), 1, "p1", 999, usecase.RoleMember)
	require.ErrorIs(t, err, usecase.ErrUserNotFound)
}

func TestRemoveMember_PreventsRemovingLastOwner(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	view, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)
	_ = view

	err = svc.RemoveMember(context.Background(), 1, "p1", 1)
	require.True(t, errors.Is(err, usecase.ErrCannotRemoveLastOwner))
}

func TestUpdateMemberRole_PreventsDemotingLastOwner(t *testing.T) {
	store := newFakeStore()
	svc := usecase.NewProjectService(store)
	_, err := svc.Create(context.Background(), 1, usecase.CreateProjectRequest{Name: "P1", Code: "p1"})
	require.NoError(t, err)

	err = svc.UpdateMemberRole(context.Background(), 1, "p1", 1, usecase.RoleMember)
	require.ErrorIs(t, err, usecase.ErrCannotRemoveLastOwner)
}
```

- [ ] **Step 3.2: 跑测试确认全失败（service 还未实现）**

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/project/usecase/...`
Expected: FAIL — `undefined: usecase.NewProjectService`.

- [ ] **Step 3.3: 实现 ProjectService**

文件: `server/service/project/usecase/project.go`

```go
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
```

- [ ] **Step 3.4: 跑测试确认全通过**

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/project/usecase/... -v`
Expected: PASS — 所有 13 个 Test* 函数全部 OK。

- [ ] **Step 3.5: 提交**

```bash
git add server/service/project/usecase/project.go server/service/project/usecase/project_test.go
git commit -m "$(cat <<'EOF'
新增 ProjectService：项目 CRUD 与成员管理业务逻辑

Constraint: 仅项目 owner 可写操作；阻止移除最后一位 owner
Confidence: high
Scope-risk: narrow
Tested: go test ./service/project/usecase/...
EOF
)"
```

---

### Task 4: GORM 持久化实现 + 集成测试

**Files:**
- Create: `server/service/project/infra/persistence/project_store_gorm.go`
- Create: `server/service/project/infra/persistence/testdb_test.go`
- Create: `server/service/project/infra/persistence/project_store_gorm_test.go`

- [ ] **Step 4.1: 写 testdb 启动辅助**

文件: `server/service/project/infra/persistence/testdb_test.go`

```go
package persistence

import (
	"os"
	"testing"

	cloudModel "optimus/server/model/cloud"
	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"
)

func openProjectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_assets").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_accounts").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS user_projects").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS projects").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS system_users").Error)

	require.NoError(t, db.AutoMigrate(
		&systemModel.SystemUser{},
		&projectModel.Project{},
		&projectModel.UserProject{},
		&cloudModel.CloudAccount{},
	))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, id uint, username string) {
	t.Helper()
	u := systemModel.SystemUser{
		ID:       id,
		UUID:     "uuid-" + username,
		Username: username,
		Source:   "local",
		Status:   "active",
	}
	require.NoError(t, db.Create(&u).Error)
}
```

- [ ] **Step 4.2: 写持久化测试（先失败）**

文件: `server/service/project/infra/persistence/project_store_gorm_test.go`

```go
package persistence

import (
	"context"
	"testing"

	projectModel "optimus/server/model/project"
	"optimus/server/service/project/usecase"

	"github.com/stretchr/testify/require"
)

func TestCreate_PersistsOwnerMembership(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)

	p := &projectModel.Project{Name: "Prod", Code: "prod", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))
	require.NotZero(t, p.ID)
	require.Equal(t, uint(1), p.OwnerID)

	m, err := store.FindMember(context.Background(), 1, p.ID)
	require.NoError(t, err)
	require.Equal(t, usecase.RoleOwner, m.ProjectRole)
}

func TestCreate_DuplicateCodeReturnsCodeTaken(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)

	require.NoError(t, store.Create(context.Background(), &projectModel.Project{Name: "P1", Code: "prod", Status: "active"}, 1))
	err := store.Create(context.Background(), &projectModel.Project{Name: "P2", Code: "prod", Status: "active"}, 1)
	require.ErrorIs(t, err, usecase.ErrProjectCodeTaken)
}

func TestListMembers_JoinsSystemUsers(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)

	p := &projectModel.Project{Name: "Prod", Code: "prod", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))
	require.NoError(t, store.AddOrReplaceMember(context.Background(), projectModel.UserProject{
		UserID: 2, ProjectID: p.ID, ProjectRole: usecase.RoleMember,
	}))

	members, err := store.ListMembers(context.Background(), p.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)

	roles := map[string]string{}
	for _, m := range members {
		roles[m.Username] = m.ProjectRole
	}
	require.Equal(t, usecase.RoleOwner, roles["alice"])
	require.Equal(t, usecase.RoleMember, roles["bob"])
}

func TestListAccessibleProjectIDs_ReturnsOnlyMemberships(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)

	p1 := &projectModel.Project{Name: "P1", Code: "p1", Status: "active"}
	p2 := &projectModel.Project{Name: "P2", Code: "p2", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p1, 1))
	require.NoError(t, store.Create(context.Background(), p2, 2))

	ids, err := store.ListAccessibleProjectIDs(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []uint{p1.ID}, ids)
}

func TestUserExists(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	store := NewGormProjectStore(db)

	ok, err := store.UserExists(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = store.UserExists(context.Background(), 999)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestCountOwners(t *testing.T) {
	db := openProjectTestDB(t)
	seedUser(t, db, 1, "alice")
	seedUser(t, db, 2, "bob")
	store := NewGormProjectStore(db)

	p := &projectModel.Project{Name: "P", Code: "p", Status: "active"}
	require.NoError(t, store.Create(context.Background(), p, 1))

	n, err := store.CountOwners(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	require.NoError(t, store.AddOrReplaceMember(context.Background(), projectModel.UserProject{
		UserID: 2, ProjectID: p.ID, ProjectRole: usecase.RoleOwner,
	}))
	n, err = store.CountOwners(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
}
```

- [ ] **Step 4.3: 跑测试确认失败**

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/project/infra/persistence/...`
Expected: FAIL — `undefined: NewGormProjectStore`.

- [ ] **Step 4.4: 实现 GormProjectStore**

文件: `server/service/project/infra/persistence/project_store_gorm.go`

```go
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
	res := s.db.WithContext(ctx).
		Model(&projectModel.Project{}).
		Where("code = ?", code).
		Updates(fields)
	if res.Error != nil {
		return projectModel.Project{}, res.Error
	}
	if res.RowsAffected == 0 {
		return projectModel.Project{}, usecase.ErrProjectNotFound
	}
	return s.FindByCode(ctx, code)
}

func (s *GormProjectStore) Delete(ctx context.Context, code string) error {
	res := s.db.WithContext(ctx).Where("code = ?", code).Delete(&projectModel.Project{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return usecase.ErrProjectNotFound
	}
	return nil
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
	res := s.db.WithContext(ctx).
		Model(&projectModel.UserProject{}).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Update("project_role", role)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return usecase.ErrMemberNotFound
	}
	return nil
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
		Select("user_projects.user_id, system_users.username, system_users.display_name, user_projects.project_role").
		Joins("JOIN system_users ON system_users.id = user_projects.user_id").
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
		Table("system_users").
		Where("id = ? AND deleted_at IS NULL", userID).
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
```

- [ ] **Step 4.5: 跑持久化测试确认通过（需 MySQL）**

Run: `cd /home/lilin/trae-project/Optimus/server && OPTIMUS_TEST_MYSQL_DSN="$OPTIMUS_TEST_MYSQL_DSN" go test ./service/project/infra/persistence/... -v`
Expected: PASS 全部 6 个测试（若无 MySQL 则全部 SKIP，编译仍要通过）。

如果无 MySQL，至少跑：
Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/project/infra/persistence/...`
Expected: 编译通过、测试 SKIP。

- [ ] **Step 4.6: 提交**

```bash
git add server/service/project/infra/persistence/
git commit -m "$(cat <<'EOF'
新增 GormProjectStore：项目与成员持久化实现

Constraint: ListMembers JOIN system_users；CountCloudAccounts 对 SP2 前 cloud_accounts 缺列容错
Confidence: high
Scope-risk: narrow
Tested: go test ./service/project/infra/persistence/... (skip without MySQL DSN)
EOF
)"
```

---

### Task 5: HTTP Handlers - Project CRUD

**Files:**
- Create: `server/api/v1/project/handlers.go`
- Create: `server/api/v1/project/project.go`

- [ ] **Step 5.1: 实现 handlers 装配类型**

文件: `server/api/v1/project/handlers.go`

```go
package project

import (
	"errors"
	"net/http"
	"strconv"

	projectUsecase "optimus/server/service/project/usecase"

	"github.com/gin-gonic/gin"
)

type ProjectHandlers struct {
	svc *projectUsecase.ProjectService
}

type ProjectHandlerDeps struct {
	Service *projectUsecase.ProjectService
}

func NewProjectHandlers(deps ProjectHandlerDeps) *ProjectHandlers {
	return &ProjectHandlers{svc: deps.Service}
}

func writeProjectServiceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": "project service is not available"})
}

func writeProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projectUsecase.ErrProjectNotFound),
		errors.Is(err, projectUsecase.ErrMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, projectUsecase.ErrProjectCodeTaken),
		errors.Is(err, projectUsecase.ErrProjectHasAccounts),
		errors.Is(err, projectUsecase.ErrCannotRemoveLastOwner):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, projectUsecase.ErrInvalidProjectCode),
		errors.Is(err, projectUsecase.ErrInvalidProjectName),
		errors.Is(err, projectUsecase.ErrInvalidProjectRole),
		errors.Is(err, projectUsecase.ErrUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, projectUsecase.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
	}
}

func projectActorID(c *gin.Context) uint {
	v, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if id, ok := v.(uint); ok {
		return id
	}
	return 0
}

func projectParseUintParam(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid " + name})
		return 0, false
	}
	return uint(n), true
}
```

- [ ] **Step 5.2: 实现 Project 5 个 HTTP 方法**

文件: `server/api/v1/project/project.go`

```go
package project

import (
	"net/http"

	projectUsecase "optimus/server/service/project/usecase"

	"github.com/gin-gonic/gin"
)

func (h *ProjectHandlers) ListProjects(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	views, err := h.svc.List(c.Request.Context(), projectActorID(c))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

func (h *ProjectHandlers) CreateProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectUsecase.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project payload"})
		return
	}
	view, err := h.svc.Create(c.Request.Context(), projectActorID(c), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *ProjectHandlers) GetProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	view, err := h.svc.Get(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *ProjectHandlers) UpdateProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectUsecase.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project payload"})
		return
	}
	view, err := h.svc.Update(c.Request.Context(), projectActorID(c), c.Param("code"), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *ProjectHandlers) DeleteProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), projectActorID(c), c.Param("code")); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}
```

- [ ] **Step 5.3: 验证编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean.

- [ ] **Step 5.4: 提交**

```bash
git add server/api/v1/project/handlers.go server/api/v1/project/project.go
git commit -m "$(cat <<'EOF'
新增 Project 模块 HTTP handlers（5 个 CRUD 路由）

Constraint: 错误映射到 401/403/404/409/400/500；actor 通过 userID context 传入
Confidence: high
Scope-risk: narrow
Tested: go build ./...
EOF
)"
```

---

### Task 6: HTTP Handlers - Member 管理

**Files:**
- Create: `server/api/v1/project/member.go`

- [ ] **Step 6.1: 实现 Member 4 个 HTTP 方法**

文件: `server/api/v1/project/member.go`

```go
package project

import (
	"net/http"

	projectUsecase "optimus/server/service/project/usecase"

	"github.com/gin-gonic/gin"
)

func (h *ProjectHandlers) ListMembers(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	members, err := h.svc.ListMembers(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": members})
}

func (h *ProjectHandlers) AddMember(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectUsecase.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid member payload"})
		return
	}
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "userId is required"})
		return
	}
	if err := h.svc.AddMember(c.Request.Context(), projectActorID(c), c.Param("code"), req.UserID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

func (h *ProjectHandlers) UpdateMemberRole(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	var req projectUsecase.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid member payload"})
		return
	}
	if err := h.svc.UpdateMemberRole(c.Request.Context(), projectActorID(c), c.Param("code"), userID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

func (h *ProjectHandlers) RemoveMember(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), projectActorID(c), c.Param("code"), userID); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}
```

- [ ] **Step 6.2: 验证编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean.

- [ ] **Step 6.3: 提交**

```bash
git add server/api/v1/project/member.go
git commit -m "$(cat <<'EOF'
新增 Project 成员管理 HTTP handlers（4 个路由）

Confidence: high
Scope-risk: narrow
Tested: go build ./...
EOF
)"
```

---

### Task 7: 路由注册

**Files:**
- Create: `server/router/project/enter.go`
- Modify: `server/router/enter.go`

- [ ] **Step 7.1: 实现路由注册**

文件: `server/router/project/enter.go`

```go
package project

import (
	projectApi "optimus/server/api/v1/project"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitProjectRouter(rg *gin.RouterGroup, h *projectApi.ProjectHandlers) {
	group := rg.Group("projects")
	{
		group.GET("", h.ListProjects)
		group.POST("", h.CreateProject)
		group.GET(":code", h.GetProject)
		group.PUT(":code", h.UpdateProject)
		group.DELETE(":code", h.DeleteProject)

		group.GET(":code/members", h.ListMembers)
		group.POST(":code/members", h.AddMember)
		group.PATCH(":code/members/:userId", h.UpdateMemberRole)
		group.DELETE(":code/members/:userId", h.RemoveMember)
	}
}
```

- [ ] **Step 7.2: 把 Project 注册到 RouterGroup**

修改 `server/router/enter.go`。先读取看现状（应该有 `Cloud cloud.RouterGroup`），在它后面加：

```go
import (
	// 现有 imports
	"optimus/server/router/cloud"
	"optimus/server/router/project"
	"optimus/server/router/system"
)

type RouterGroup struct {
	System  system.RouterGroup
	Cloud   cloud.RouterGroup
	Project project.RouterGroup
}
```

具体改动取决于当前文件内容；保持已有结构，只插入 `Project project.RouterGroup` 字段与 import 行。

- [ ] **Step 7.3: 验证编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean.

- [ ] **Step 7.4: 提交**

```bash
git add server/router/project/ server/router/enter.go
git commit -m "$(cat <<'EOF'
注册 Project 模块路由（9 个端点挂到 privateGroup）

Confidence: high
Scope-risk: narrow
Tested: go build ./...
EOF
)"
```

---

### Task 8: 模块装配 + Initialize 接入

**Files:**
- Create: `server/initialize/project/module.go`
- Modify: `server/initialize/router.go`
- Modify: `server/initialize/startup.go`

- [ ] **Step 8.1: 实现 NewProjectModule**

文件: `server/initialize/project/module.go`

```go
package project

import (
	projectApi "optimus/server/api/v1/project"
	"optimus/server/config"
	"optimus/server/service/project/infra/persistence"
	projectUsecase "optimus/server/service/project/usecase"

	"gorm.io/gorm"
)

type ProjectModule struct {
	Handlers *projectApi.ProjectHandlers
}

func NewProjectModule(db *gorm.DB, _ config.Config) *ProjectModule {
	if db == nil {
		return &ProjectModule{Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{})}
	}
	store := persistence.NewGormProjectStore(db)
	svc := projectUsecase.NewProjectService(store)
	return &ProjectModule{
		Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{Service: svc}),
	}
}
```

- [ ] **Step 8.2: 在 router.go 接入 ProjectModule**

修改 `server/initialize/router.go`，在 cloudModule 同级新增：

```go
// 新增 imports
import (
	cloudInit "optimus/server/initialize/cloud"
	projectInit "optimus/server/initialize/project"
	// 既有 imports
)

// 已有的 var 区段附近：
var newProjectModule = projectInit.NewProjectModule

// 在 Routers() 函数里，cloudModule 同级新增：
projectModule := newProjectModule(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)
// ...在 privateGroup 注册区块加：
routerGroup.Project.InitProjectRouter(privateGroup, projectModule.Handlers)
```

具体替换以现有文件结构为准。保持与 `newCloudModule` 平行。

- [ ] **Step 8.3: 在 startup.go 接入迁移与种子**

修改 `server/initialize/startup.go`，在 `cloudService.SeedCloudData` 之后追加：

```go
import (
	projectService "optimus/server/service/project"
	// 既有 imports
)

// 在 SeedCloudData 后面：
if err := projectService.AutoMigrateProject(db); err != nil {
	logDependencyInitFailure("project", "auto migrate", err)
	return
}
if err := projectService.SeedProjectData(db); err != nil {
	logDependencyInitFailure("project", "seed data", err)
	return
}
```

- [ ] **Step 8.4: 全量编译 + 全量测试**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...`
Expected: build clean, all packages PASS (persistence 测试 SKIP 若无 MySQL)。

- [ ] **Step 8.5: 提交**

```bash
git add server/initialize/project/ server/initialize/router.go server/initialize/startup.go
git commit -m "$(cat <<'EOF'
装配 Project 模块：路由注册、迁移与权限种子接入 startup

Constraint: NewProjectModule 在 db=nil 时降级为 service unavailable handlers
Confidence: high
Scope-risk: narrow（无侵入 cloud / system 模块）
Tested: go build ./... && go test ./...
EOF
)"
```

---

## Self-Review Checklist

**Spec coverage:**
- ✓ `projects` 表（Task 1）
- ✓ `user_projects` 表（Task 1）
- ✓ 项目角色 owner/member/viewer（Task 2 常量）
- ✓ 9 个 REST 端点（Task 5/6/7）
- ✓ project.* 权限种子（Task 1 step 1.2）
- ✓ admin 角色自动授权全部 project.*（Task 1 step 1.2 assignAdminProjectPermissions）
- ✓ 创建者自动成为 owner（Task 4 GormProjectStore.Create transaction）
- ✓ 阻止移除最后一位 owner（Task 3 RemoveMember / UpdateMemberRole 测试）
- ✓ 删除项目时校验 cloud_accounts 绑定（Task 3 Delete; Task 4 CountCloudAccounts 容错列缺失）
- ✓ project code 正则 `^[a-z][a-z0-9-]{1,62}$`（Task 3 projectCodeRegex）

**Placeholder scan:** 无 TBD/TODO/"实现 X"占位；所有代码段完整可拷贝。

**Type consistency:**
- `ProjectService.Create/List/...` 签名在 Task 3 定义，Task 5 handlers 直接调用，签名一致。
- `ProjectStore` 在 Task 2 定义，Task 4 实现，签名一致。
- `RoleOwner/Member/Viewer` 常量在 Task 2 定义，Task 3/4 引用。
- `ErrProjectCodeTaken`/`ErrCannotRemoveLastOwner` 等 sentinel 在 Task 2 定义，Task 3/4/5 引用。
- `ProjectMember{UserID, Username, DisplayName, ProjectRole}` 字段在 Task 2 定义，Task 4 ListMembers 的 Select 列名一致映射。

**Scope check:** SP1 完全不动 cloud 模块；唯一与 cloud 的边界是 `CountCloudAccounts` 软读 `cloud_accounts.project_id`（容错处理 SP2 前的列缺失）。

---

## Out of Scope (deferred to SP2/SP3/SP4)

- middleware 注入 ctx.currentProjectID（SP2）
- cloud_accounts 加 project_id 列（SP2）
- 前端项目管理页（SP3）
- 顶部项目切换器（SP3）
- 云账号/资产页项目化（SP4）
