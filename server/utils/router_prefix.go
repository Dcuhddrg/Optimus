package utils

import (
	"strings"

	"optimus/server/global"
)

// DefaultRouterPrefix is used when system.router-prefix is empty.
const DefaultRouterPrefix = "/api/v1"

// RouterPrefix returns the configured route prefix with a leading slash.
func RouterPrefix() string {
	prefix := strings.TrimSpace(global.OPTIMUS_CONFIG.System.RouterPrefix)
	if prefix == "" {
		return DefaultRouterPrefix
	}
	if strings.HasPrefix(prefix, "/") {
		return prefix
	}
	return "/" + prefix
}
