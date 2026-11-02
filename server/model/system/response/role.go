package response

import (
	systemModel "optimus/server/model/system"
)

type RoleListResult struct {
	Items []systemModel.SystemRole `json:"items"`
	Total int                      `json:"total"`
}
