package usecase

import "strconv"

const (
	defaultAdminUsername = "admin"
	defaultUserStatus    = "enabled"
	defaultRoleStatus    = "enabled"
)

func resourceID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
