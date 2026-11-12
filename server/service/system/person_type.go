package system

import (
	"strings"
	"time"

	systemModel "optimus/server/model/system"
)

// Person types (design 3.6). External accounts expire, need a sponsor, and
// cannot hold the admin role, approval roles or project ownership.
const (
	PersonEmployee = "employee"
	PersonExternal = "external"
)

// splitTypeValues reads the stored comma-separated external type list.
func splitTypeValues(stored string) []string {
	out := []string{}
	for _, v := range strings.Split(stored, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// joinTypeValues normalizes a list for storage: trimmed, without empty or
// repeated values (case-insensitive) and without commas inside a value.
func joinTypeValues(values []string) string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		v = strings.TrimSpace(strings.ReplaceAll(v, ",", ""))
		if v == "" || seen[strings.ToLower(v)] {
			continue
		}
		seen[strings.ToLower(v)] = true
		out = append(out, v)
	}
	return strings.Join(out, ",")
}

// personTypeOf classifies a directory entry: external when any of its
// employee type values is in the configured list, otherwise employee.
func personTypeOf(entry LDAPUser, cfg systemModel.SystemLDAPConfig) string {
	for _, external := range splitTypeValues(cfg.ExternalTypeValues) {
		for _, value := range entry.EmployeeTypes {
			if strings.EqualFold(strings.TrimSpace(value), external) {
				return PersonExternal
			}
		}
	}
	return PersonEmployee
}

// personTypeOrDefault treats an unset column (older rows) as employee.
func personTypeOrDefault(personType string) string {
	if personType == "" {
		return PersonEmployee
	}
	return personType
}

// isRestrictedForExternal reports roles an external account cannot hold:
// the admin role and approval roles. Custom roles are not restricted.
func isRestrictedForExternal(code string, approvable bool) bool {
	return code == AdminRoleCode || approvable
}

// effectiveRoles are the roles authorization uses. An external account
// missing its expiry or sponsor has none; otherwise its admin and approval
// roles are ignored (design 3.6). Employees keep all their roles.
func effectiveRoles(user systemModel.SystemUser, roles []systemModel.SystemRoleSummary) []systemModel.SystemRoleSummary {
	if personTypeOrDefault(user.PersonType) != PersonExternal {
		return roles
	}
	out := []systemModel.SystemRoleSummary{}
	if user.ExpiresAt == nil || user.SponsorID == nil {
		return out
	}
	for _, role := range roles {
		if !isRestrictedForExternal(role.Code, role.Approvable) {
			out = append(out, role)
		}
	}
	return out
}

// endOfDay turns a YYYY-MM-DD date into the last instant of that day in
// server time: the account is valid through the whole day.
func endOfDay(date string) (time.Time, error) {
	day, err := time.ParseInLocation(time.DateOnly, date, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return day.AddDate(0, 0, 1).Add(-time.Millisecond), nil
}

// expiresOn is the last valid day of an expiry, as YYYY-MM-DD.
func expiresOn(expiresAt *time.Time) *string {
	if expiresAt == nil {
		return nil
	}
	day := expiresAt.In(time.Local).Format(time.DateOnly)
	return &day
}

// accountExpired reports whether an external account is past its expiry.
// Expiry is independent of Status: it never changes it, and extending the
// expiry restores the account at once.
func accountExpired(personType string, expiresAt *time.Time, now time.Time) bool {
	return personTypeOrDefault(personType) == PersonExternal && expiresAt != nil && now.After(*expiresAt)
}
