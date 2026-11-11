package response

import (
	systemModel "optimus/server/model/system"
)

type RoleListResult struct {
	Items []systemModel.SystemRole `json:"items"`
	Total int                      `json:"total"`
}

// RoleListItem is a role with its permission codes, for the role editor.
type RoleListItem struct {
	systemModel.SystemRole
	PermissionCodes []string `json:"permissionCodes"`
}
