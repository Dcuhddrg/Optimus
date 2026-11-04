package ldap

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

// ErrUserNotFound is returned when the user filter does not match exactly one entry.
var ErrUserNotFound = errors.New("ldap user not found")

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
}

// User is the directory identity returned by a successful authentication.
type User struct {
	Username    string
	DisplayName string
	Email       string
}

// Client is the LDAP capability used by services.
type Client interface {
	Test(ctx context.Context, cfg Config) error
	Authenticate(ctx context.Context, cfg Config, username, password string) (User, error)
}

// NewLDAP builds a client per call, like GVA's upload.NewOss.
func NewLDAP() Client {
	return &GoLDAPClient{Timeout: 5 * time.Second}
}

type GoLDAPClient struct {
	Timeout time.Duration
}

func (c *GoLDAPClient) Test(_ context.Context, cfg Config) error {
	conn, err := ldap.DialURL(cfg.URL, ldap.DialWithDialer(&net.Dialer{Timeout: c.timeout()}))
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		return err
	}

	request := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		int(c.timeout().Seconds()),
		false,
		replaceLDAPUsername(cfg.UserFilter, "*"),
		[]string{cfg.UsernameAttr},
		nil,
	)
	_, err = conn.Search(request)
	return err
}

func (c *GoLDAPClient) Authenticate(_ context.Context, cfg Config, username string, password string) (User, error) {
	conn, err := ldap.DialURL(cfg.URL, ldap.DialWithDialer(&net.Dialer{Timeout: c.timeout()}))
	if err != nil {
		return User{}, err
	}
	defer conn.Close()

	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		return User{}, err
	}

	attrs := []string{cfg.UsernameAttr, cfg.DisplayNameAttr, cfg.EmailAttr}
	request := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		int(c.timeout().Seconds()),
		false,
		replaceLDAPUsername(cfg.UserFilter, username),
		attrs,
		nil,
	)
	result, err := conn.Search(request)
	if err != nil {
		return User{}, err
	}
	if len(result.Entries) != 1 {
		return User{}, ErrUserNotFound
	}

	entry := result.Entries[0]
	if err := conn.Bind(entry.DN, password); err != nil {
		return User{}, err
	}

	return User{
		Username:    firstLDAPAttr(entry, cfg.UsernameAttr, username),
		DisplayName: firstLDAPAttr(entry, cfg.DisplayNameAttr, username),
		Email:       firstLDAPAttr(entry, cfg.EmailAttr, ""),
	}, nil
}

func (c *GoLDAPClient) timeout() time.Duration {
	if c.Timeout <= 0 {
		return 5 * time.Second
	}
	return c.Timeout
}

func replaceLDAPUsername(filter string, username string) string {
	return strings.ReplaceAll(filter, "{username}", ldap.EscapeFilter(username))
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
