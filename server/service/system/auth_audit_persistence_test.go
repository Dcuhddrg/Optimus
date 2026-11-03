package system_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"optimus/server/global"
	model "optimus/server/model/system"
	request "optimus/server/model/system/request"
	service "optimus/server/service/system"
)

func TestFailedLoginAuditUsesUnverifiedAttempt(t *testing.T) {
	db := openExternalPersistenceDB(t)
	svc := service.NewAuthService(nil, nil)
	_, err := svc.Login(context.Background(), request.LoginRequest{Username: "unknown\naccount", Password: "PASSWORD_SECRET"}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Zero(t, audit.ActorUserID)
	require.Empty(t, audit.ActorUsername)
	require.Empty(t, audit.ResourceID)
	require.NotContains(t, audit.Detail, "PASSWORD_SECRET")
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, float64(1), detail["version"])
	require.Equal(t, "unknown account", detail["attemptedUsername"])
}

func TestDisabledLoginAuditDoesNotClaimActor(t *testing.T) {
	db := openExternalPersistenceDB(t)
	require.NoError(t, db.Create(&model.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "test-user", Username: "admin", Source: "local", Status: "disabled"}).Error)
	_, err := service.NewAuthService(nil, nil).Login(context.Background(), request.LoginRequest{Username: "admin", Password: "PASSWORD_SECRET"}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrUserDisabled)
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Zero(t, audit.ActorUserID)
	require.Empty(t, audit.ActorUsername)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, float64(7), detail["targetUserId"])
}

func TestFailedLoginAuditBoundsAttempt(t *testing.T) {
	db := openExternalPersistenceDB(t)
	_, err := service.NewAuthService(nil, nil).Login(context.Background(), request.LoginRequest{Username: strings.Repeat("界", 300)}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.LessOrEqual(t, len([]rune(detail["attemptedUsername"].(string))), 128)
}
