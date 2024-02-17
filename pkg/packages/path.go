package packages

import (
	"path"
)

func init() {
	Packages.Insert("path", PackageMap{
		"Base":          path.Base,
		"Clean":         path.Clean,
		"Dir":           path.Dir,
		"ErrBadPattern": path.ErrBadPattern,
		"Ext":           path.Ext,
		"IsAbs":         path.IsAbs,
		"Join":          path.Join,
		"Match":         path.Match,
		"Split":         path.Split,
	})
}
