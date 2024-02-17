package packages

import (
	"io/ioutil"
)

func init() {
	Packages.Insert("io/ioutil", PackageMap{
		"ReadAll":   ioutil.ReadAll,
		"ReadDir":   ioutil.ReadDir,
		"ReadFile":  ioutil.ReadFile,
		"WriteFile": ioutil.WriteFile,
	})
}
