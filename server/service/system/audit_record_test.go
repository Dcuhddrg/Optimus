package system

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"log"
	"optimus/server/global"
	"testing"
)

type failedAuditWriter struct{}

func (failedAuditWriter) WriteAudit(context.Context, AuditEvent) error {
	return errors.New("password=SECRET token=SECRET SQL=SECRET")
}
func TestAuditRecordFallbackDoesNotLeak(t *testing.T) {
	old := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = nil
	t.Cleanup(func() { global.OPTIMUS_LOG = old })
	out := log.Writer()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(out) })
	recordAudit(context.Background(), failedAuditWriter{}, AuditEvent{Action: "role.permissions.update", ResourceType: "role", ResourceID: "3", Result: "success", Detail: "SECRET"})
	require.Contains(t, buf.String(), "audit write failed")
	require.Contains(t, buf.String(), "role.permissions.update")
	require.NotContains(t, buf.String(), "SECRET")
}
func TestAuditDetailsDropsNonWhitelistedFields(t *testing.T) {
	raw := auditDetails("updated", map[string]auditChange{"password": {Before: "SECRET", After: "NEW_SECRET"}, "token": {Before: "TOKEN", After: "NEW_TOKEN"}, "status": {Before: "enabled", After: "disabled"}})
	require.NotContains(t, raw, "SECRET")
	require.NotContains(t, raw, "TOKEN")
	require.Contains(t, raw, "status")
}
func TestAuditDetailsEncodingFailureStillProducesValidDetail(t *testing.T) {
	raw := auditDetails("updated", map[string]auditChange{"status": {Before: make(chan int), After: "enabled"}})
	var detail auditRecordDetail
	require.NoError(t, json.Unmarshal([]byte(raw), &detail))
	require.Equal(t, 1, detail.Version)
	require.Equal(t, "detail_unavailable", detail.Reason)
}
func TestAuditSetChangeCanonicalizesAndDoesNotModifyInput(t *testing.T) {
	before := []uint{3, 1, 1}
	after := []uint{4, 3, 4}
	change := auditSetChange(before, after)
	require.Equal(t, []uint{3, 1, 1}, before)
	require.Equal(t, []uint{4, 3, 4}, after)
	require.Equal(t, []uint{1, 3}, change.Before)
	require.Equal(t, []uint{3, 4}, change.After)
	require.Equal(t, []uint{4}, change.Added)
	require.Equal(t, []uint{1}, change.Removed)
}
