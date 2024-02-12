package packages

import (
	"os/signal"
)

func init() {
	Packages["os/signal"] = map[string]any{
		"Notify": signal.Notify,
		"Stop":   signal.Stop,
	}
}
