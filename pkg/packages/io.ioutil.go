package packages

import (
	"io/ioutil"
)

func init() {
	Packages["io/ioutil"] = map[string]any{
		"ReadAll":   ioutil.ReadAll,
		"ReadDir":   ioutil.ReadDir,
		"ReadFile":  ioutil.ReadFile,
		"WriteFile": ioutil.WriteFile,
	}
}
