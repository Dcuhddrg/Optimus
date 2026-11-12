package ldap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

var (
	// ErrUserNotFound is returned when the login filter matches no entry.
	ErrUserNotFound = errors.New("ldap user not found")
	// ErrDuplicateUsername is returned when the login filter matches more than
	// one entry: OpenLDAP does not keep uid unique without slapo-unique.
	ErrDuplicateUsername = errors.New("ldap username matches more than one entry")
	// ErrInvalidCredentials is returned when the user's bind is rejected.
	ErrInvalidCredentials = errors.New("ldap invalid credentials")
	// ErrMissingIdentity is returned when the entry has no identity attribute.
	ErrMissingIdentity = errors.New("ldap entry has no directory identity")
	// ErrDuplicateIdentity is returned when the identity value is not unique
	// under the base DN.
	ErrDuplicateIdentity = errors.New("ldap directory identity is not unique")
	// ErrMissingVersion is returned when neither entryCSN nor modifyTimestamp
	// can be read, so the bind cannot be tied to the entry that was searched.
	ErrMissingVersion = errors.New("ldap entry has no readable version")
	// ErrEntryChanged is returned when the entry changed between the search,
	// the bind and the re-read.
	ErrEntryChanged = errors.New("ldap entry changed during login")
	// ErrNoUsersInScope is returned by Test when the sync scope is empty.
	ErrNoUsersInScope = errors.New("no users in the sync scope; check the base DN and the filter")
	// ErrSampleMissingIdentity is returned by Test when a sampled entry has no
	// readable identity attribute.
	ErrSampleMissingIdentity = errors.New("cannot read the directory identity; check the attribute name and the bind account's read access")
)

// Config is the connection and attribute mapping used for one LDAP call.
type Config struct {
	URL             string
	BaseDN          string
	BindDN          string
	BindPassword    string
	UserFilter      string
	UsernameAttr    string
	DisplayNameAttr string
	EmailAttr       string
	DepartmentAttr  string
	// EmployeeTypeAttr is read with all its values (employeeType is
	// multi-valued); empty means not read.
	EmployeeTypeAttr string
	IDAttr           string
}

// User is a directory entry read for one user. Empty attributes stay empty:
// LDAP is the source of truth for these fields.
type User struct {
	DN          string
	Username    string
	DisplayName string
	Email       string
	Department  string
	// EmployeeTypes are all values of EmployeeTypeAttr.
	EmployeeTypes []string
	ExternalID    string
	// Version is entryCSN (OpenLDAP) or modifyTimestamp; rename and move
	// update it.
	Version string
	// CreatedAt is createTimestamp, read by the sync scope reads only; nil
	// when unreadable.
	CreatedAt *time.Time
}

// Client is the LDAP capability used by services.
type Client interface {
	// Test checks the connection and samples the sync scope.
	Test(ctx context.Context, cfg Config) error
	// Authenticate verifies the password and returns the entry whose identity
	// was read before and after the bind.
	Authenticate(ctx context.Context, cfg Config, username, password string) (User, error)
	// LookupUsername reads the entry for username with the service account.
	LookupUsername(ctx context.Context, cfg Config, username string) (User, error)
	// SearchScope reads every entry of the sync scope.
	SearchScope(ctx context.Context, cfg Config) ([]User, error)
	// LookupExternalID re-reads one sync-scope entry by its identity.
	LookupExternalID(ctx context.Context, cfg Config, externalID string) (User, error)
	// ExistsInDirectory looks the identity up across the whole directory.
	ExistsInDirectory(ctx context.Context, cfg Config, externalID string) (bool, error)
}

// NewLDAP builds a client per call, like GVA's upload.NewOss.
func NewLDAP() Client {
	return &GoLDAPClient{Timeout: 5 * time.Second}
}

// GoLDAPClient talks to a directory with go-ldap.
type GoLDAPClient struct {
	Timeout time.Duration
}

const (
	attrEntryCSN        = "entryCSN"
	attrModifyTimestamp = "modifyTimestamp"
)

func (c *GoLDAPClient) dial(cfg Config) (*ldap.Conn, error) {
	conn, err := ldap.DialURL(cfg.URL, ldap.DialWithDialer(&net.Dialer{Timeout: c.timeout()}))
	if err != nil {
		return nil, err
	}
	conn.SetTimeout(c.timeout())
	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (c *GoLDAPClient) Test(_ context.Context, cfg Config) error {
	syncFilter, err := ValidateLoginFilter(cfg.UserFilter, cfg.UsernameAttr)
	if err != nil {
		return err
	}
	conn, err := c.dial(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	// A size limit of 1 samples one entry; more matches are expected and
	// reported by the server as "size limit exceeded".
	result, err := conn.Search(c.searchRequest(cfg.BaseDN, ldap.ScopeWholeSubtree, 1, syncFilter, []string{cfg.UsernameAttr, cfg.IDAttr}))
	if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultSizeLimitExceeded) {
		return err
	}
	if result == nil || len(result.Entries) == 0 {
		return ErrNoUsersInScope
	}
	if result.Entries[0].GetAttributeValue(cfg.IDAttr) == "" {
		return ErrSampleMissingIdentity
	}
	return nil
}

func (c *GoLDAPClient) Authenticate(_ context.Context, cfg Config, username string, password string) (User, error) {
	if _, err := ValidateLoginFilter(cfg.UserFilter, cfg.UsernameAttr); err != nil {
		return User{}, err
	}
	conn, err := c.dial(cfg)
	if err != nil {
		return User{}, err
	}
	defer conn.Close()

	before, err := c.findByUsername(conn, cfg, username)
	if err != nil {
		return User{}, err
	}
	if err := conn.Bind(before.DN, password); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}
	// The password is verified; only now report identity problems, so they
	// are not disclosed to someone without the password.
	if before.ExternalID == "" {
		return User{}, ErrMissingIdentity
	}
	if before.Version == "" {
		return User{}, ErrMissingVersion
	}

	// LDAP has no bind-and-read: re-read the same DN as the service account.
	// The entry must still carry the identity and version seen before the
	// bind, otherwise the DN was replaced, renamed or moved in between.
	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		return User{}, err
	}
	after, err := conn.Search(c.searchRequest(before.DN, ldap.ScopeBaseObject, 1, "(objectClass=*)", []string{cfg.IDAttr, attrEntryCSN, attrModifyTimestamp}))
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			return User{}, ErrEntryChanged
		}
		return User{}, err
	}
	if len(after.Entries) != 1 || after.Entries[0].GetAttributeValue(cfg.IDAttr) != before.ExternalID || entryVersion(after.Entries[0]) != before.Version {
		return User{}, ErrEntryChanged
	}

	// Second line of defence after the whitelist: the identity must be unique.
	byID, err := conn.Search(c.searchRequest(cfg.BaseDN, ldap.ScopeWholeSubtree, 2, fmt.Sprintf("(%s=%s)", cfg.IDAttr, ldap.EscapeFilter(before.ExternalID)), []string{cfg.IDAttr}))
	if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultSizeLimitExceeded) {
		return User{}, err
	}
	if err != nil || len(byID.Entries) != 1 {
		return User{}, ErrDuplicateIdentity
	}
	return before, nil
}

func (c *GoLDAPClient) LookupUsername(_ context.Context, cfg Config, username string) (User, error) {
	if _, err := ValidateLoginFilter(cfg.UserFilter, cfg.UsernameAttr); err != nil {
		return User{}, err
	}
	conn, err := c.dial(cfg)
	if err != nil {
		return User{}, err
	}
	defer conn.Close()
	return c.findByUsername(conn, cfg, username)
}

// findByUsername searches with the login filter. A size limit of 2 tells a
// unique match from a duplicated username.
func (c *GoLDAPClient) findByUsername(conn *ldap.Conn, cfg Config, username string) (User, error) {
	result, err := conn.Search(c.searchRequest(cfg.BaseDN, ldap.ScopeWholeSubtree, 2, replaceLDAPUsername(cfg.UserFilter, username), userAttributes(cfg)))
	if ldap.IsErrorWithCode(err, ldap.LDAPResultSizeLimitExceeded) {
		return User{}, ErrDuplicateUsername
	}
	if err != nil {
		return User{}, err
	}
	switch len(result.Entries) {
	case 0:
		return User{}, ErrUserNotFound
	case 1:
	default:
		return User{}, ErrDuplicateUsername
	}
	return entryUser(result.Entries[0], cfg), nil
}

func (c *GoLDAPClient) searchRequest(base string, scope int, sizeLimit int, filter string, attrs []string) *ldap.SearchRequest {
	return ldap.NewSearchRequest(base, scope, ldap.NeverDerefAliases, sizeLimit, int(c.timeout().Seconds()), false, filter, attrs, nil)
}

// userAttributes lists the mapped attributes plus the operational ones,
// which a server only returns when asked for by name.
func userAttributes(cfg Config) []string {
	attrs := []string{attrEntryCSN, attrModifyTimestamp}
	for _, attr := range []string{cfg.UsernameAttr, cfg.DisplayNameAttr, cfg.EmailAttr, cfg.DepartmentAttr, cfg.EmployeeTypeAttr, cfg.IDAttr} {
		if attr != "" {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}

func entryVersion(entry *ldap.Entry) string {
	if csn := entry.GetAttributeValue(attrEntryCSN); csn != "" {
		return csn
	}
	return entry.GetAttributeValue(attrModifyTimestamp)
}

func (c *GoLDAPClient) timeout() time.Duration {
	if c.Timeout <= 0 {
		return 5 * time.Second
	}
	return c.Timeout
}

func replaceLDAPUsername(filter string, username string) string {
	return strings.ReplaceAll(filter, usernamePlaceholder, ldap.EscapeFilter(username))
}

func firstLDAPAttr(entry *ldap.Entry, attr string, fallback string) string {
	if attr == "" {
		return fallback
	}
	value := entry.GetAttributeValue(attr)
	if value == "" {
		return fallback
	}
	return value
}
