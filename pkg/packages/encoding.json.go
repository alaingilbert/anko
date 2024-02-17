package packages

import (
	"encoding/json"
)

func init() {
	Packages.Insert("encoding/json", PackageMap{
		"Marshal":   json.Marshal,
		"Unmarshal": json.Unmarshal,
	})
}
