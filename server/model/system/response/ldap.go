package response

type LDAPConfigResult struct {
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	BaseDN          string `json:"baseDn"`
	BindDN          string `json:"bindDn"`
	BindPassword    string `json:"bindPassword,omitempty"`
	HasBindPassword bool   `json:"hasBindPassword"`
	UserFilter      string `json:"userFilter"`
	UsernameAttr    string `json:"usernameAttr"`
	DisplayNameAttr string `json:"displayNameAttr"`
	EmailAttr       string `json:"emailAttr"`
}

type LDAPTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
