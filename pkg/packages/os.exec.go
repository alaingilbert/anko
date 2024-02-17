package packages

import (
	"os/exec"
)

func init() {
	Packages.Insert("os/exec", PackageMap{
		"ErrNotFound": exec.ErrNotFound,
		"LookPath":    exec.LookPath,
		"Command":     exec.Command,
	})
	PackageTypes.Insert("os/exec", PackageMap{
		"Cmd": exec.Cmd{},
	})
}
