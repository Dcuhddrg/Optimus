package response

type ProjectMember struct {
	UserID      uint   `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	ProjectRole string `json:"projectRole"`
}
