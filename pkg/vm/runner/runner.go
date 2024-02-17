package runner

import (
	"context"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/utils/ratelimitanything"
	"github.com/alaingilbert/anko/pkg/utils/stateCh"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"reflect"
	"sync"
	"sync/atomic"
)

// MapLocker we need to lock map operations MapIndex/SetMapIndex/mapIter.Next
// in the VM to avoid application crash if the script uses a map in a concurrent manner.
type MapLocker struct{ sync.Mutex }

func (m *MapLocker) Lock()   { m.Mutex.Lock() }
func (m *MapLocker) Unlock() { m.Mutex.Unlock() }

type VmParams struct {
	ctx              context.Context
	rvCh             chan Result
	stats            *Stats
	doNotProtectMaps bool
	mapMutex         *MapLocker
	pause            *stateCh.StateCh
	rateLimit        *ratelimitanything.RateLimitAnything
	Validate         bool
	has              map[any]bool
	ValidateLater    map[string]ast.Stmt
}

func NewVmParams(ctx context.Context,
	rvCh chan Result,
	stats *Stats,
	doNotProtectMaps bool,
	mapMutex *MapLocker,
	pause *stateCh.StateCh,
	rateLimit *ratelimitanything.RateLimitAnything,
	validate bool,
	has map[any]bool,
	validateLater map[string]ast.Stmt,
) *VmParams {
	return &VmParams{
		ctx:              ctx,
		rvCh:             rvCh,
		stats:            stats,
		doNotProtectMaps: doNotProtectMaps,
		mapMutex:         mapMutex,
		pause:            pause,
		rateLimit:        rateLimit,
		Validate:         validate,
		has:              has,
		ValidateLater:    validateLater,
	}
}

func Run(vmp *VmParams, env envPkg.IEnv, stmt ast.Stmt) (reflect.Value, error) {
	return runSingleStmt(vmp, env, stmt)
}

type Result struct {
	Value reflect.Value
	Error error
}

type Stats struct {
	Cycles int64
}

func incrCycle(vmp *VmParams) error {
	// make sure script is not stopped
	select {
	case <-vmp.ctx.Done():
		return ErrInterrupt
	default:
	}
	// if script is NOT paused, `<-vmp.pause.Wait()` will return right away
	select {
	case <-vmp.pause.Wait():
	case <-vmp.ctx.Done():
		return ErrInterrupt
	}
	// halt here if we need to throttle the script
	rateLimit := vmp.rateLimit
	if rateLimit != nil {
		select {
		case <-rateLimit.GetWithContext(vmp.ctx):
		case <-vmp.ctx.Done():
			return ErrInterrupt
		}
	}
	atomic.AddInt64(&vmp.stats.Cycles, 1)
	return nil
}