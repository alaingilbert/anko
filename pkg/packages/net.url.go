//go:build !appengine

package packages

import (
	"net/url"
)

func init() {
	Packages.Insert("net/url", PackageMap{
		"QueryEscape":     url.QueryEscape,
		"QueryUnescape":   url.QueryUnescape,
		"Parse":           url.Parse,
		"ParseRequestURI": url.ParseRequestURI,
		"User":            url.User,
		"UserPassword":    url.UserPassword,
		"ParseQuery":      url.ParseQuery,
	})
	PackageTypes.Insert("net/url", PackageMap{
		"Error":  url.Error{},
		"URL":    url.URL{},
		"Values": url.Values{},
	})
}
