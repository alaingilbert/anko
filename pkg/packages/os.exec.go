package packages

import (
	"os/exec"
)

func init() {
	Packages.Insert("os/exec", map[string]any{
		"ErrNotFound": exec.ErrNotFound,
		"LookPath":    exec.LookPath,
		"Command":     exec.Command,
	})
}
