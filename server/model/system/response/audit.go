package response

import (
	commonResponse "optimus/server/model/common/response"
	systemModel "optimus/server/model/system"
)

// AuditListData is the HTTP data payload for migrated audit queries.
// AuditPage remains the internal Service/Store contract.
type AuditListData commonResponse.PageResult[systemModel.AuditLogEntry]
