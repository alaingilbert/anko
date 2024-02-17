package packages

import "github.com/alaingilbert/mtx"

type PackageMap map[string]any

var (
	// Packages is a where all the packages are stored, so they can be imported when wanted
	Packages = mtx.NewRWMapPtr[string, PackageMap](nil)

	//Packages = make(map[string]map[string]any, 16)

	// PackageTypes is a where all the package types are stored, so they can be imported when wanted
	PackageTypes = mtx.NewRWMapPtr[string, PackageMap](nil)
	//PackageTypes = make(map[string]map[string]any, 4)
)
