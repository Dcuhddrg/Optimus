package cloud

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"optimus/server/global"
	cloudModel "optimus/server/model/cloud"
	"optimus/server/service"
	"testing"

	cloudService "optimus/server/service/cloud"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setCloudServicesForTest(t *testing.T, group cloudService.ServiceGroup) {
	t.Helper()
	previousGroup := service.ServiceGroupApp.CloudServiceGroup
	previousAccount, previousAsset := accountService, assetService
	t.Cleanup(func() {
		service.ServiceGroupApp.CloudServiceGroup = previousGroup
		accountService, assetService = previousAccount, previousAsset
	})
	service.ServiceGroupApp.CloudServiceGroup = group
	RefreshServiceAliases()
}

func TestAccountHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setCloudServicesForTest(t, cloudService.ServiceGroup{})

	engine := gin.New()
	engine.GET("/cloud/accounts", (&AccountApi{}).ListAccounts)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/accounts", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestAssetHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setCloudServicesForTest(t, cloudService.ServiceGroup{})

	engine := gin.New()
	engine.GET("/cloud/assets", (&AssetApi{}).ListAssets)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/assets", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestAccountHandlerCreateRequiresCurrentProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountSvc, err := cloudService.NewCloudAccountService(make([]byte, 32))
	require.NoError(t, err)
	setCloudServicesForTest(t, cloudService.ServiceGroup{CloudAccountService: *accountSvc})

	engine := gin.New()
	engine.POST("/cloud/accounts", (&AccountApi{}).CreateAccount)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cloud/accounts", bytes.NewBufferString(`{"name":"prod","provider":"aliyun","accountId":"123","credentialType":"ak_sk","credential":{"access_key_id":"ak"}}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAccountHandlerCreateUsesCurrentProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "test:test@tcp(127.0.0.1:3306)/optimus",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	previousDB := global.OPTIMUS_DB
	t.Cleanup(func() {
		global.OPTIMUS_DB = previousDB
		require.NoError(t, sqlDB.Close())
	})
	var projectID uint
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:capture_project_id", func(tx *gorm.DB) {
		projectID = tx.Statement.Dest.(*cloudModel.CloudAccount).ProjectID
	}))
	global.OPTIMUS_DB = db
	accountSvc, err := cloudService.NewCloudAccountService(make([]byte, 32))
	require.NoError(t, err)
	setCloudServicesForTest(t, cloudService.ServiceGroup{CloudAccountService: *accountSvc})

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(utils.ProjectScopeContextKey, utils.ProjectScope{CurrentProjectID: 7, AccessibleProjectIDs: []uint{7}})
		c.Next()
	})
	engine.POST("/cloud/accounts", (&AccountApi{}).CreateAccount)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cloud/accounts", bytes.NewBufferString(`{"name":"prod","provider":"aliyun","accountId":"123","credentialType":"ak_sk","credential":{"access_key_id":"ak"}}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, uint(7), projectID)
}
