package request

import (
	commonRequest "optimus/server/model/common/request"
	"time"
)

// AuditSearch describes HTTP query inputs. API parsing preserves the existing
// permissive handling of invalid optional IDs and RFC3339 timestamps.
type AuditSearch struct {
	commonRequest.PageInfo
	ActorUserID  *uint      `json:"actorUserId" form:"actorUserId"`
	Action       *string    `json:"action" form:"action"`
	ResourceType *string    `json:"resourceType" form:"resourceType"`
	Result       *string    `json:"result" form:"result"`
	StartTime    *time.Time `json:"startTime" form:"startTime"`
	EndTime      *time.Time `json:"endTime" form:"endTime"`
	Keyword      *string    `json:"keyword" form:"keyword"`
}
