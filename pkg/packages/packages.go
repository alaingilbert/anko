package packages

var (
	// Packages is a where all the packages are stored so they can be imported when wanted
	Packages = make(map[string]map[string]any, 16)
	// PackageTypes is a where all the package types are stored so they can be imported when wanted
	PackageTypes = make(map[string]map[string]any, 4)
)
