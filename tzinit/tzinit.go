package tzinit

import (
	"os"
)

func init() {
	err := os.Setenv("TZ", "Europe/Istanbul")
	if err != nil {
		return
	}
}
