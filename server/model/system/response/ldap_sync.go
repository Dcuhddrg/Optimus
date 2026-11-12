package response

import "time"

// LDAPSyncAccount is the Optimus side of a sync preview row.
type LDAPSyncAccount struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email"`
	Department  string    `json:"department"`
	ExternalID  string    `json:"externalId"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	Roles       []string  `json:"roles"`
}

// LDAPSyncEntry is the directory side of a sync preview row.
type LDAPSyncEntry struct {
	DN          string     `json:"dn"`
	Username    string     `json:"username"`
	DisplayName string     `json:"displayName"`
	Email       string     `json:"email"`
	Department  string     `json:"department"`
	ExternalID  string     `json:"externalId"`
	CreatedAt   *time.Time `json:"createdAt"`
}

// LDAPSyncFieldChange is one profile field an update changes.
type LDAPSyncFieldChange struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// LDAPSyncItem is one row of a sync preview. Category is one of create,
// update, autoDisable, unconfirmable, outOfScope, pendingBinding, conflict,
// disabledInfo, unidentifiable.
type LDAPSyncItem struct {
	Category   string                `json:"category"`
	Reason     string                `json:"reason,omitempty"`
	Account    *LDAPSyncAccount      `json:"account,omitempty"`
	Entry      *LDAPSyncEntry        `json:"entry,omitempty"`
	Changes    []LDAPSyncFieldChange `json:"changes,omitempty"`
	Suspicious bool                  `json:"suspicious,omitempty"`
	// RestrictedRoles and OwnedProjects are set when an update turns an
	// employee into an external person: those roles stop applying at once;
	// the admin should remove them and hand the projects to an employee.
	RestrictedRoles []string `json:"restrictedRoles,omitempty"`
	OwnedProjects   []string `json:"ownedProjects,omitempty"`
}

// LDAPSyncPreview is a read-only sync plan. Executing it needs PreviewID.
type LDAPSyncPreview struct {
	PreviewID string    `json:"previewId"`
	ExpiresAt time.Time `json:"expiresAt"`
	// DisablesBlocked: some entries could not be identified, so this sync
	// disables nobody.
	DisablesBlocked bool `json:"disablesBlocked"`
	// ConfigChanged: base DN, filter or identity attribute changed since the
	// last successful sync, so every disable candidate needs the admin.
	ConfigChanged bool           `json:"configChanged"`
	Items         []LDAPSyncItem `json:"items"`
}

// LDAPSyncOutcome is what execution did for one person. Outcome is done,
// skipped, error or unprocessed.
type LDAPSyncOutcome struct {
	Action    string `json:"action"`
	AccountID uint   `json:"accountId,omitempty"`
	Username  string `json:"username"`
	Outcome   string `json:"outcome"`
	Reason    string `json:"reason,omitempty"`
}

// LDAPSyncResult reports an execution. Status is completed or partial
// (stopped at a directory error; committed people stay committed).
type LDAPSyncResult struct {
	Status   string            `json:"status"`
	Outcomes []LDAPSyncOutcome `json:"outcomes"`
}
