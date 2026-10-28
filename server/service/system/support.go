package system

import "strconv"

func resourceID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
