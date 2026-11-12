package request

type LDAPConfigRequest struct {
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	BaseDN          string `json:"baseDn"`
	BindDN          string `json:"bindDn"`
	BindPassword    string `json:"bindPassword"`
	UserFilter      string `json:"userFilter"`
	UsernameAttr    string `json:"usernameAttr"`
	DisplayNameAttr string `json:"displayNameAttr"`
	EmailAttr       string `json:"emailAttr"`
	IDAttr          string `json:"idAttr"`
	// DepartmentAttr nil keeps the saved value; "" stops syncing departments.
	DepartmentAttr *string `json:"departmentAttr"`
	// EmployeeTypeAttr "" keeps the saved value (employeeType by default).
	EmployeeTypeAttr string `json:"employeeTypeAttr"`
	// ExternalTypeValues nil keeps the saved list; an empty list means
	// nobody is external.
	ExternalTypeValues *[]string `json:"externalTypeValues"`
}

type LDAPTestRequest struct {
	URL                  string `json:"url"`
	BaseDN               string `json:"baseDn"`
	BindDN               string `json:"bindDn"`
	BindPassword         string `json:"bindPassword"`
	UseSavedBindPassword bool   `json:"useSavedBindPassword"`
	UserFilter           string `json:"userFilter"`
	UsernameAttr         string `json:"usernameAttr"`
	DisplayNameAttr      string `json:"displayNameAttr"`
	EmailAttr            string `json:"emailAttr"`
	IDAttr               string `json:"idAttr"`
	DepartmentAttr       string `json:"departmentAttr"`
}

// LDAPSyncExecuteRequest executes a sync preview with the admin's choices.
// Each list holds account IDs from the preview: Disable for outOfScope and
// unconfirmable rows, Bind for pendingBinding rows confirmed as the same
// person, Separate for pendingBinding or conflict rows confirmed as another
// person.
type LDAPSyncExecuteRequest struct {
	PreviewID string `json:"previewId" binding:"required"`
	Disable   []uint `json:"disable"`
	Bind      []uint `json:"bind"`
	Separate  []uint `json:"separate"`
}
