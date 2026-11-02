package request

type RoleRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleStatusRequest struct {
	Status string `json:"status"`
}

type UpdateRolePermissionsRequest struct {
	PermissionCodes []string `json:"permissionCodes"`
}
