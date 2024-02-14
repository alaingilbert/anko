package packages

import (
	"runtime"
)

func init() {
	Packages.Insert("runtime", map[string]any{
		"GC":         runtime.GC,
		"GOARCH":     runtime.GOARCH,
		"GOMAXPROCS": runtime.GOMAXPROCS,
		"GOOS":       runtime.GOOS,
		"GOROOT":     runtime.GOROOT,
		"Version":    runtime.Version,
	})
}
