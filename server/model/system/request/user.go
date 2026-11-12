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

// UpdateUserExternalRequest sets an external account's expiry and sponsor
// (design 3.6). Both are replaced; null clears one. ExpiresOn is a date
// (YYYY-MM-DD): the account expires when that day ends, server time.
type UpdateUserExternalRequest struct {
	ExpiresOn *string `json:"expiresOn"`
	SponsorID *uint   `json:"sponsorId"`
}
