package packages

import (
	"errors"
)

func init() {
	Packages.Insert("errors", map[string]any{
		"New": errors.New,
	})
}
