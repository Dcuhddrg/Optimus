package response

type LDAPConfigResult struct {
	Enabled            bool     `json:"enabled"`
	URL                string   `json:"url"`
	BaseDN             string   `json:"baseDn"`
	BindDN             string   `json:"bindDn"`
	BindPassword       string   `json:"bindPassword,omitempty"`
	HasBindPassword    bool     `json:"hasBindPassword"`
	UserFilter         string   `json:"userFilter"`
	UsernameAttr       string   `json:"usernameAttr"`
	DisplayNameAttr    string   `json:"displayNameAttr"`
	EmailAttr          string   `json:"emailAttr"`
	IDAttr             string   `json:"idAttr"`
	DepartmentAttr     string   `json:"departmentAttr"`
	EmployeeTypeAttr   string   `json:"employeeTypeAttr"`
	ExternalTypeValues []string `json:"externalTypeValues"`
	// BoundUsers counts accounts bound to a directory identity; while it is
	// non-zero, IDAttr cannot change.
	BoundUsers int64 `json:"boundUsers"`
}

type LDAPTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
