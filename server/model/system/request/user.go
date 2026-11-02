package request

type UpdateUserRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status"`
}

type UpdateUserRolesRequest struct {
	RoleIDs []uint `json:"roleIds"`
}
