package packages

import (
	"sync"
)

func init() {
	Packages["sync"] = map[string]any{
		"NewCond": sync.NewCond,
	}
	PackageTypes["sync"] = map[string]any{
		"Cond":      sync.Cond{},
		"Mutex":     sync.Mutex{},
		"Once":      sync.Once{},
		"Pool":      sync.Pool{},
		"RWMutex":   sync.RWMutex{},
		"WaitGroup": sync.WaitGroup{},
	}
	syncGo19()
}
