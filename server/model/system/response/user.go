package response

type SystemUserSummary struct {
	ID          uint   `json:"id"`
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	Bound       bool   `json:"bound"`
	Source      string `json:"source"`
	Status      string `json:"status"`
	// PersonType is employee or external. ExpiresOn (YYYY-MM-DD, the last
	// valid day) and SponsorID apply to external accounts only.
	PersonType string  `json:"personType"`
	ExpiresOn  *string `json:"expiresOn"`
	SponsorID  *uint   `json:"sponsorId"`
}

type UserListResult struct {
	Items []UserListItem `json:"items"`
	Total int            `json:"total"`
}

type UserListItem struct {
	SystemUserSummary
	Roles []string `json:"roles"`
	// Sponsor is the sponsor's username; SponsorInvalid is set when the
	// sponsor no longer exists, is disabled or is not an employee. Expired
	// is set when an external account is past its expiry.
	Sponsor        string `json:"sponsor,omitempty"`
	SponsorInvalid bool   `json:"sponsorInvalid,omitempty"`
	Expired        bool   `json:"expired,omitempty"`
	// For an external account: held roles that do not apply (admin and
	// approval roles) and projects it still owns, until an administrator
	// removes or hands them over (design 3.6).
	RestrictedRoles []string `json:"restrictedRoles,omitempty"`
	OwnedProjects   []string `json:"ownedProjects,omitempty"`
}
