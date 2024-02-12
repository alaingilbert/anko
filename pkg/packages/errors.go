package packages

import (
	"errors"
)

func init() {
	Packages["errors"] = map[string]any{
		"New": errors.New,
	}
}
