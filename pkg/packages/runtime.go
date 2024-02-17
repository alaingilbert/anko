package packages

import (
	"runtime"
)

func init() {
	Packages.Insert("runtime", PackageMap{
		"GC":         runtime.GC,
		"GOARCH":     runtime.GOARCH,
		"GOMAXPROCS": runtime.GOMAXPROCS,
		"GOOS":       runtime.GOOS,
		"GOROOT":     runtime.GOROOT,
		"Version":    runtime.Version,
	})
}
