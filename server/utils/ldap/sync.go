package ldap

import (
	"context"
	"errors"
	"fmt"

	ber "github.com/go-asn1-ber/asn1-ber"
	ldap "github.com/go-ldap/ldap/v3"
)

const (
	attrCreateTimestamp = "createTimestamp"
	attrNamingContexts  = "namingContexts"
	syncPageSize        = 500
)

// ErrNoNamingContexts is returned when the RootDSE lists no naming context,
// so absence cannot be confirmed across the whole directory.
var ErrNoNamingContexts = errors.New("ldap RootDSE has no namingContexts")

// SearchScope reads every entry in the sync scope (base DN and the derived
// sync filter) with paged results, so the server's size limit does not cut
// the list short. Any error aborts the whole read.
func (c *GoLDAPClient) SearchScope(_ context.Context, cfg Config) ([]User, error) {
	syncFilter, err := ValidateLoginFilter(cfg.UserFilter, cfg.UsernameAttr)
	if err != nil {
		return nil, err
	}
	conn, err := c.dial(cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	attrs := append(userAttributes(cfg), attrCreateTimestamp)
	result, err := conn.SearchWithPaging(c.searchRequest(cfg.BaseDN, ldap.ScopeWholeSubtree, 0, syncFilter, attrs), syncPageSize)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0, len(result.Entries))
	for _, entry := range result.Entries {
		users = append(users, entryUser(entry, cfg))
	}
	return users, nil
}

// LookupExternalID re-reads one entry of the sync scope by its identity.
func (c *GoLDAPClient) LookupExternalID(_ context.Context, cfg Config, externalID string) (User, error) {
	syncFilter, err := ValidateLoginFilter(cfg.UserFilter, cfg.UsernameAttr)
	if err != nil {
		return User{}, err
	}
	conn, err := c.dial(cfg)
	if err != nil {
		return User{}, err
	}
	defer conn.Close()
	filter := fmt.Sprintf("(&%s(%s=%s))", syncFilter, cfg.IDAttr, ldap.EscapeFilter(externalID))
	attrs := append(userAttributes(cfg), attrCreateTimestamp)
	result, err := conn.Search(c.searchRequest(cfg.BaseDN, ldap.ScopeWholeSubtree, 2, filter, attrs))
	if ldap.IsErrorWithCode(err, ldap.LDAPResultSizeLimitExceeded) {
		return User{}, ErrDuplicateIdentity
	}
	if err != nil {
		return User{}, err
	}
	switch len(result.Entries) {
	case 0:
		return User{}, ErrUserNotFound
	case 1:
		return entryUser(result.Entries[0], cfg), nil
	default:
		return User{}, ErrDuplicateIdentity
	}
}

// ExistsInDirectory confirms absence: it searches the identity under every
// naming context of the RootDSE, without the login filter, so entries moved
// out of the base DN or out of the filter are still found.
func (c *GoLDAPClient) ExistsInDirectory(_ context.Context, cfg Config, externalID string) (bool, error) {
	conn, err := c.dial(cfg)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	root, err := conn.Search(c.searchRequest("", ldap.ScopeBaseObject, 1, "(objectClass=*)", []string{attrNamingContexts}))
	if err != nil {
		return false, err
	}
	var contexts []string
	if len(root.Entries) == 1 {
		contexts = root.Entries[0].GetAttributeValues(attrNamingContexts)
	}
	if len(contexts) == 0 {
		return false, ErrNoNamingContexts
	}
	filter := fmt.Sprintf("(%s=%s)", cfg.IDAttr, ldap.EscapeFilter(externalID))
	for _, base := range contexts {
		result, err := conn.Search(c.searchRequest(base, ldap.ScopeWholeSubtree, 1, filter, []string{cfg.IDAttr}))
		if ldap.IsErrorWithCode(err, ldap.LDAPResultSizeLimitExceeded) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		if len(result.Entries) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func entryUser(entry *ldap.Entry, cfg Config) User {
	user := User{
		DN:          entry.DN,
		Username:    firstLDAPAttr(entry, cfg.UsernameAttr, ""),
		DisplayName: firstLDAPAttr(entry, cfg.DisplayNameAttr, ""),
		Email:       firstLDAPAttr(entry, cfg.EmailAttr, ""),
		Department:  firstLDAPAttr(entry, cfg.DepartmentAttr, ""),
		ExternalID:  firstLDAPAttr(entry, cfg.IDAttr, ""),
		Version:     entryVersion(entry),
	}
	if cfg.EmployeeTypeAttr != "" {
		user.EmployeeTypes = entry.GetAttributeValues(cfg.EmployeeTypeAttr)
	}
	if raw := entry.GetAttributeValue(attrCreateTimestamp); raw != "" {
		if created, err := ber.ParseGeneralizedTime([]byte(raw)); err == nil {
			user.CreatedAt = &created
		}
	}
	return user
}
