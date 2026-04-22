package core

import "fmt"

func Banner(version string, addr int) string {
	return fmt.Sprintf(
		"Optimus backend\nversion: %s\nlistening on: :%d\n",
		version,
		addr,
	)
}
