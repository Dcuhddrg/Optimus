package request

type RoleRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Approvable: nil keeps the current flag on update, false on create.
	Approvable *bool `json:"approvable"`
}

type UpdateRolePermissionsRequest struct {
	PermissionCodes []string `json:"permissionCodes"`
}
