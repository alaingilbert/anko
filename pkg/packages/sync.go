package packages

import (
	"sync"
)

func init() {
	Packages.Insert("sync", map[string]any{
		"NewCond": sync.NewCond,
	})
	PackageTypes.Insert("sync", map[string]any{
		"Cond":      sync.Cond{},
		"Mutex":     sync.Mutex{},
		"Once":      sync.Once{},
		"Pool":      sync.Pool{},
		"RWMutex":   sync.RWMutex{},
		"WaitGroup": sync.WaitGroup{},
		"Map":       sync.Map{},
	})
}
