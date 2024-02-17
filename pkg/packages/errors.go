package packages

import (
	"errors"
)

func init() {
	Packages.Insert("errors", PackageMap{
		"New": errors.New,
	})
}
