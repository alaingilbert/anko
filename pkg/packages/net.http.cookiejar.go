package packages

import (
	"net/http/cookiejar"
)

func init() {
	Packages.Insert("net/http/cookiejar", map[string]any{
		"New": cookiejar.New,
	})
	PackageTypes.Insert("net/http/cookiejar", map[string]any{
		"Options": cookiejar.Options{},
	})
}
