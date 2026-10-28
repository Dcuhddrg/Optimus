package cloud

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudService "optimus/server/service/cloud"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testAccountStore struct {
	rows map[uint]cloudService.AccountRow
	next uint
}

func (s *testAccountStore) Create(_ context.Context, projectID uint, row cloudService.AccountRow) (cloudService.AccountRow, error) {
	s.next++
	row.ID = s.next
	row.ProjectID = projectID
	row.Status = "active"
	if row.Tags == nil {
		row.Tags = map[string]string{}
	}
	s.rows[row.ID] = row
	return row, nil
}

func (s *testAccountStore) Get(_ context.Context, _ []uint, id uint) (cloudService.AccountRow, error) {
	row, ok := s.rows[id]
	if !ok {
		return cloudService.AccountRow{}, cloudService.ErrAccountNotFound
	}
	return row, nil
}

func (s *testAccountStore) GetForSync(_ context.Context, id uint) (cloudService.AccountRow, error) {
	return s.Get(context.Background(), nil, id)
}

func (s *testAccountStore) List(context.Context, []uint, []cloudService.TagFilter) ([]cloudService.AccountRow, error) {
	return nil, nil
}

func (s *testAccountStore) Update(context.Context, []uint, uint, *string, *string) (cloudService.AccountRow, error) {
	return cloudService.AccountRow{}, nil
}

func (s *testAccountStore) UpdateTags(_ context.Context, _ []uint, id uint, tags map[string]string) (cloudService.AccountRow, error) {
	row, ok := s.rows[id]
	if !ok {
		return cloudService.AccountRow{}, cloudService.ErrAccountNotFound
	}
	row.Tags = tags
	s.rows[id] = row
	return row, nil
}

func (s *testAccountStore) Delete(context.Context, []uint, uint) error {
	return nil
}

func (s *testAccountStore) UpdateSyncStatus(context.Context, uint, *time.Time, string) error {
	return nil
}

func TestAccountHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCloudHandlers(CloudHandlerDeps{})

	engine := gin.New()
	engine.GET("/cloud/accounts", h.Account.ListAccounts)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/accounts", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestAssetHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCloudHandlers(CloudHandlerDeps{})

	engine := gin.New()
	engine.GET("/cloud/assets", h.Asset.ListAssets)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/assets", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestAccountHandlerCreateRequiresCurrentProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountSvc, err := cloudService.NewCloudAccountService(&testAccountStore{rows: map[uint]cloudService.AccountRow{}}, make([]byte, 32))
	require.NoError(t, err)
	h := NewCloudHandlers(CloudHandlerDeps{Account: accountSvc})

	engine := gin.New()
	engine.POST("/cloud/accounts", h.Account.CreateAccount)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cloud/accounts", bytes.NewBufferString(`{"name":"prod","provider":"aliyun","accountId":"123","credentialType":"ak_sk","credential":{"access_key_id":"ak"}}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAccountHandlerCreateUsesCurrentProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &testAccountStore{rows: map[uint]cloudService.AccountRow{}}
	accountSvc, err := cloudService.NewCloudAccountService(store, make([]byte, 32))
	require.NoError(t, err)
	h := NewCloudHandlers(CloudHandlerDeps{Account: accountSvc})

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(utils.ProjectScopeContextKey, utils.ProjectScope{CurrentProjectID: 7, AccessibleProjectIDs: []uint{7}})
		c.Next()
	})
	engine.POST("/cloud/accounts", h.Account.CreateAccount)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cloud/accounts", bytes.NewBufferString(`{"name":"prod","provider":"aliyun","accountId":"123","credentialType":"ak_sk","credential":{"access_key_id":"ak"}}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, uint(7), store.rows[1].ProjectID)
}
