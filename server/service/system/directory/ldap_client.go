package directory

import (
	"context"
	"net"
	"strings"
	"time"

	systemService "optimus/server/service/system"

	ldap "github.com/go-ldap/ldap/v3"
)

type GoLDAPClient struct {
	Timeout time.Duration
}

func NewGoLDAPClient() *GoLDAPClient {
	return &GoLDAPClient{Timeout: 5 * time.Second}
}

func (c *GoLDAPClient) Test(_ context.Context, cfg systemService.LDAPConnectionConfig) error {
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

func (c *GoLDAPClient) Authenticate(_ context.Context, cfg systemService.LDAPConnectionConfig, username string, password string) (systemService.LDAPUser, error) {
	conn, err := ldap.DialURL(cfg.URL, ldap.DialWithDialer(&net.Dialer{Timeout: c.timeout()}))
	if err != nil {
		return systemService.LDAPUser{}, err
	}
	defer conn.Close()

	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		return systemService.LDAPUser{}, err
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
		return systemService.LDAPUser{}, err
	}
	if len(result.Entries) != 1 {
		return systemService.LDAPUser{}, systemService.ErrInvalidCredentials
	}

	entry := result.Entries[0]
	if err := conn.Bind(entry.DN, password); err != nil {
		return systemService.LDAPUser{}, err
	}

	return systemService.LDAPUser{
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
