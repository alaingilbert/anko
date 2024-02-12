package packages

import (
	"net/http/cookiejar"
)

func init() {
	Packages["net/http/cookiejar"] = map[string]any{
		"New": cookiejar.New,
	}
	PackageTypes["net/http/cookiejar"] = map[string]any{
		"Options": cookiejar.Options{},
	}
}
