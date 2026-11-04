package cloud

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"optimus/server/global"
	cloudModel "optimus/server/model/cloud"
	"testing"

	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setCloudReadyForTest sets the inputs cloudReady reads: the database handle
// (only when db is non-nil) and the configured credential key.
func setCloudReadyForTest(t *testing.T, db *gorm.DB, key string) {
	t.Helper()
	previousDB, previousKey := global.OPTIMUS_DB, global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey
	t.Cleanup(func() {
		global.OPTIMUS_DB = previousDB
		global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey = previousKey
	})
	global.OPTIMUS_DB = db
	global.OPTIMUS_CONFIG.Cloud.CredentialEncryptionKey = key
}

var validTestCredentialKey = string(make([]byte, 32))

func TestAccountHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setCloudReadyForTest(t, nil, validTestCredentialKey)

	engine := gin.New()
	engine.GET("/cloud/accounts", (&AccountApi{}).ListAccounts)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/accounts", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":7`)
}

func TestAccountHandlerListReturnsServiceUnavailableWithInvalidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setCloudReadyForTest(t, &gorm.DB{}, "tooshort")

	engine := gin.New()
	engine.GET("/cloud/accounts", (&AccountApi{}).ListAccounts)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/cloud/accounts", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":7`)
}

func TestAssetHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setCloudReadyForTest(t, nil, validTestCredentialKey)

	engine := gin.New()
	engine.GET("/cloud/assets", (&AssetApi{}).ListAssets)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/assets", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":7`)
}

func TestAccountHandlerCreateRequiresCurrentProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setCloudReadyForTest(t, &gorm.DB{}, validTestCredentialKey)

	engine := gin.New()
	engine.POST("/cloud/accounts", (&AccountApi{}).CreateAccount)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cloud/accounts", bytes.NewBufferString(`{"name":"prod","provider":"aliyun","accountId":"123","credentialType":"ak_sk","credential":{"access_key_id":"ak"}}`))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":7`)
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
	setCloudReadyForTest(t, db, validTestCredentialKey)

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

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":0`)
	require.Equal(t, uint(7), projectID)
}
