package response

type LoginResult struct {
	Token           string            `json:"token"`
	User            SystemUserSummary `json:"user"`
	Roles           []string          `json:"roles"`
	MenuPermissions []string          `json:"menuPermissions"`
}

type CurrentUserResult struct {
	User            SystemUserSummary `json:"user"`
	Roles           []string          `json:"roles"`
	MenuPermissions []string          `json:"menuPermissions"`
}
