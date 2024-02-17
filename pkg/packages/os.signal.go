package packages

import (
	"os/signal"
)

func init() {
	Packages.Insert("os/signal", PackageMap{
		"Notify": signal.Notify,
		"Stop":   signal.Stop,
	})
}
