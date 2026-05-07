package system

type AuditMeta struct {
	IP        string
	UserAgent string
}

type AuditEvent struct {
	ActorUserID  uint
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	IP           string
	UserAgent    string
	Detail       string
}
