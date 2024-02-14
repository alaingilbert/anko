package packages

import "github.com/alaingilbert/mtx"

var (
	// Packages is a where all the packages are stored, so they can be imported when wanted
	Packages = mtx.NewRWMapPtr[string, map[string]any](nil)

	//Packages = make(map[string]map[string]any, 16)

	// PackageTypes is a where all the package types are stored, so they can be imported when wanted
	PackageTypes = mtx.NewRWMapPtr[string, map[string]any](nil)
	//PackageTypes = make(map[string]map[string]any, 4)
)
