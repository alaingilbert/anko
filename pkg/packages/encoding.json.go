package packages

import (
	"encoding/json"
)

func init() {
	Packages.Insert("encoding/json", map[string]any{
		"Marshal":   json.Marshal,
		"Unmarshal": json.Unmarshal,
	})
}
