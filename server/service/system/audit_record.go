package system

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"slices"
	"unicode"

	"go.uber.org/zap"
	"optimus/server/global"
)

type auditWriter interface {
	WriteAudit(context.Context, AuditEvent) error
}

type auditChange struct {
	Before  any `json:"before"`
	After   any `json:"after"`
	Added   any `json:"added,omitempty"`
	Removed any `json:"removed,omitempty"`
}

type auditRecordDetail struct {
	Version           int                    `json:"version"`
	Reason            string                 `json:"reason"`
	Changes           map[string]auditChange `json:"changes,omitempty"`
	AttemptedUsername string                 `json:"attemptedUsername,omitempty"`
	TargetUserID      uint                   `json:"targetUserId,omitempty"`
}

// Auditing is best effort: do not turn a committed business operation into an API failure.
// Raw store errors may contain SQL and credentials; only their type is safe to log here.
func recordAudit(ctx context.Context, writer auditWriter, event AuditEvent) {
	if err := writer.WriteAudit(ctx, event); err != nil {
		if logger, ok := global.OPTIMUS_LOG.(*zap.Logger); ok && logger != nil {
			logger.Error("audit write failed", zap.String("action", event.Action), zap.String("resource_type", event.ResourceType), zap.String("resource_id", event.ResourceID), zap.String("result", event.Result), zap.String("error_type", fmt.Sprintf("%T", err)))
		} else {
			log.Printf("audit write failed action=%q resource_type=%q resource_id=%q result=%q error_type=%T", event.Action, event.ResourceType, event.ResourceID, event.Result, err)
		}
	}
}

func auditDetails(reason string, changes map[string]auditChange) string {
	safe := make(map[string]auditChange)
	for key, change := range changes {
		switch key {
		case "status", "roleIds", "permissionCodes", "enabled", "connectionChanged", "mappingChanged", "bindPasswordChanged":
			if !reflect.DeepEqual(change.Before, change.After) {
				safe[key] = change
			}
		}
	}
	data, err := json.Marshal(auditRecordDetail{Version: 1, Reason: reason, Changes: safe})
	if err != nil {
		return `{"version":1,"reason":"detail_unavailable"}`
	}
	return string(data)
}

func failedAudit(actorID uint, meta AuditMeta, action, resourceType string, targetID uint) AuditEvent {
	target := ""
	if targetID != 0 {
		target = resourceID(targetID)
	}
	return AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: action, ResourceType: resourceType, ResourceID: target, Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: auditDetails("operation_failed", nil)}
}

func loginFailure(reqUsername string, meta AuditMeta, targetID uint, reason string) AuditEvent {
	name := []rune(reqUsername)
	if len(name) > 128 {
		name = name[:128]
	}
	for i, r := range name {
		if unicode.IsControl(r) {
			name[i] = ' '
		}
	}
	data, _ := json.Marshal(auditRecordDetail{Version: 1, Reason: reason, AttemptedUsername: string(name), TargetUserID: targetID})
	event := failedAudit(0, meta, "auth.login.failure", "user", targetID)
	event.ActorUsername = "" // A failed attempt does not establish an authenticated actor.
	event.Detail = string(data)
	return event
}

func canonicalAuditSet[T cmp.Ordered](values []T) []T {
	out := make([]T, len(values))
	copy(out, values)
	slices.Sort(out)
	return slices.Compact(out)
}
func auditSetChange[T cmp.Ordered](before, after []T) auditChange {
	before = canonicalAuditSet(before)
	after = canonicalAuditSet(after)
	added, removed := make([]T, 0), make([]T, 0)
	for _, v := range after {
		if !slices.Contains(before, v) {
			added = append(added, v)
		}
	}
	for _, v := range before {
		if !slices.Contains(after, v) {
			removed = append(removed, v)
		}
	}
	return auditChange{Before: before, After: after, Added: added, Removed: removed}
}
