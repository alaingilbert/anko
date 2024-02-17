package packages

import (
	"net/http/cookiejar"
)

func init() {
	Packages.Insert("net/http/cookiejar", PackageMap{
		"New": cookiejar.New,
	})
	PackageTypes.Insert("net/http/cookiejar", PackageMap{
		"Options": cookiejar.Options{},
	})
}
