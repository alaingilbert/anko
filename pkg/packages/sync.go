package packages

import (
	"sync"
)

func init() {
	Packages.Insert("sync", PackageMap{
		"NewCond": sync.NewCond,
	})
	PackageTypes.Insert("sync", PackageMap{
		"Cond":      sync.Cond{},
		"Mutex":     sync.Mutex{},
		"Once":      sync.Once{},
		"Pool":      sync.Pool{},
		"RWMutex":   sync.RWMutex{},
		"WaitGroup": sync.WaitGroup{},
		"Map":       sync.Map{},
	})
}
