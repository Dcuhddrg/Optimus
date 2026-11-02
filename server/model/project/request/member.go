package request

type AddMemberRequest struct {
	UserID      uint   `json:"userId"`
	ProjectRole string `json:"projectRole"`
}

type UpdateMemberRoleRequest struct {
	ProjectRole string `json:"projectRole"`
}
