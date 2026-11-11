package ldap

import (
	"errors"
	"fmt"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
)

const usernamePlaceholder = "{username}"

// ErrInvalidLoginFilter is returned for a login filter whose structure would
// not give the same population for login and sync (design 3.3 D3).
var ErrInvalidLoginFilter = errors.New("login filter must contain (<username attribute>={username}) exactly once, either as the whole filter or as a direct child of the outermost (&...); it must not appear under !, |, a nested &, or a substring match")

// IdentityAttributes are the directory identity attributes Optimus accepts:
// each is generated and kept unique by the directory server.
var IdentityAttributes = []string{"entryUUID", "nsUniqueId", "ipaUniqueID"}

// IsIdentityAttribute reports whether attr is on the identity whitelist.
func IsIdentityAttribute(attr string) bool {
	for _, allowed := range IdentityAttributes {
		if allowed == attr {
			return true
		}
	}
	return false
}

// ValidateLoginFilter checks the login filter structure and returns the sync
// filter: the username item replaced by (<usernameAttr>=*), everything else
// unchanged. The replacement is structural, so the * is a presence test, not
// an escaped value.
func ValidateLoginFilter(filter string, usernameAttr string) (string, error) {
	if usernameAttr == "" || strings.Count(filter, usernamePlaceholder) != 1 {
		return "", ErrInvalidLoginFilter
	}
	root, err := ldap.CompileFilter(filter)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidLoginFilter, err)
	}
	// An escaped \7busername\7d decodes to the placeholder text too; any second
	// node carrying it would make the population ambiguous.
	if countPlaceholderNodes(root) != 1 {
		return "", ErrInvalidLoginFilter
	}
	presence := fmt.Sprintf("(%s=*)", usernameAttr)
	if isUsernameItem(root, usernameAttr) {
		return presence, nil
	}
	if root.Tag != ldap.FilterAnd {
		return "", ErrInvalidLoginFilter
	}
	var sync strings.Builder
	sync.WriteString("(&")
	found := false
	for _, child := range root.Children {
		if isUsernameItem(child, usernameAttr) {
			found = true
			sync.WriteString(presence)
			continue
		}
		part, err := ldap.DecompileFilter(child)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrInvalidLoginFilter, err)
		}
		sync.WriteString(part)
	}
	if !found {
		return "", ErrInvalidLoginFilter
	}
	sync.WriteString(")")
	return sync.String(), nil
}

func isUsernameItem(packet *ber.Packet, usernameAttr string) bool {
	if packet.Tag != ldap.FilterEqualityMatch || len(packet.Children) != 2 {
		return false
	}
	attr, _ := packet.Children[0].Value.(string)
	value, _ := packet.Children[1].Value.(string)
	return strings.EqualFold(attr, usernameAttr) && value == usernamePlaceholder
}

func countPlaceholderNodes(packet *ber.Packet) int {
	count := 0
	if value, ok := packet.Value.(string); ok && strings.Contains(value, usernamePlaceholder) {
		count++
	}
	for _, child := range packet.Children {
		count += countPlaceholderNodes(child)
	}
	return count
}
