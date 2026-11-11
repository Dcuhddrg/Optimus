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
