package packages

import (
	"os/signal"
)

func init() {
	Packages.Insert("os/signal", map[string]any{
		"Notify": signal.Notify,
		"Stop":   signal.Stop,
	})
}
