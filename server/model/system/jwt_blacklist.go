package system

import "optimus/server/global"

// JwtBlacklist stores revoked tokens until they expire, as GVA's sys_jwt_blacklist.
type JwtBlacklist struct {
	global.OPTIMUS_MODEL
	Jwt string `gorm:"type:text;comment:jwt"`
}
