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
}
