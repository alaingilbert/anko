package packages

import (
	"regexp"
)

func init() {
	Packages.Insert("regexp", PackageMap{
		"Match":            regexp.Match,
		"MatchReader":      regexp.MatchReader,
		"MatchString":      regexp.MatchString,
		"QuoteMeta":        regexp.QuoteMeta,
		"Compile":          regexp.Compile,
		"CompilePOSIX":     regexp.CompilePOSIX,
		"MustCompile":      regexp.MustCompile,
		"MustCompilePOSIX": regexp.MustCompilePOSIX,
	})
}
