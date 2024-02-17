package vm

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/packages"
	"github.com/alaingilbert/anko/pkg/parser"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/ratelimitanything"
	"github.com/alaingilbert/anko/pkg/vm/stateCh"
	"github.com/alaingilbert/mtx"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type IVM interface {
	Executor() IExecutor
	Validate(context.Context, any) error
	Has(context.Context, any, []any) ([]bool, error)

	Define(k string, v any) error
	DefineType(k string, v any) error
	AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (envPkg.IEnv, error)
}

// Compile time checks to ensure type satisfies IVM interface
var _ IVM = (*VM)(nil)

type IExecutor interface {
	Stop()
	Run(ctx context.Context, input any) (any, error)
	RunAsync(ctx context.Context, input any)
	Pause()
	Resume()
	IsPaused() bool
	GetCycles() int64

	GetEnv() envPkg.IEnv
}

var _ IExecutor = (*Executor)(nil)

type Executor struct {
	env              envPkg.IEnv                          // executor's env
	pause            *stateCh.StateCh                     // allows pause/resume of scripts
	stats            *stats                               // keep track of stmt/expr processed
	rateLimit        *ratelimitanything.RateLimitAnything // rate limit expr processed/duration
	doNotProtectMaps bool                                 // either or not to protect maps operations in the VM
	mapMutex         *mapLocker                           // locker object to protect maps
	cancel           context.CancelFunc                   // use to Stop a script
	importCore       bool                                 // either or not to import core functions in executor's env
	watchdogEnabled  bool                                 // either or not to run the watchdog
	maxEnvCount      *mtx.Mtx[int64]                      // maximum sub-env allowed before the watchdog kills the script
}

type ExecutorConfig struct {
	DoNotProtectMaps bool
	DoNotDeepCopyEnv bool
	ImportCore       bool
	WatchdogEnabled  bool
	DefineImport     bool
	RateLimit        int
	RateLimitPeriod  time.Duration
	Env              envPkg.IEnv
	MaxEnvCount      int
}

func NewExecutor(cfg *ExecutorConfig) *Executor {
	if cfg == nil {
		return nil
	}
	e := &Executor{}
	if cfg.DoNotDeepCopyEnv {
		e.env = cfg.Env
	} else {
		e.env = cfg.Env.DeepCopy()
	}
	if cfg.ImportCore {
		Import(e.env)
	}
	if cfg.DefineImport {
		DefineImport(e.env)
	}
	if cfg.WatchdogEnabled {
		e.watchdogEnabled = cfg.WatchdogEnabled
	}
	e.pause = stateCh.NewStateCh(true)
	e.stats = &stats{}
	e.doNotProtectMaps = cfg.DoNotProtectMaps
	e.importCore = cfg.ImportCore
	e.mapMutex = &mapLocker{}
	e.maxEnvCount = mtx.NewRWMtxPtr(int64(cfg.MaxEnvCount))
	if cfg.RateLimit > 0 {
		e.rateLimit = ratelimitanything.NewRateLimitAnything(int64(cfg.RateLimit), cfg.RateLimitPeriod)
	}
	return e
}

func (e *Executor) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
}

func (e *Executor) Run(ctx context.Context, input any) (any, error) {
	return e.run(ctx, input)
}

func (e *Executor) RunAsync(ctx context.Context, input any) {
	e.runAsync(ctx, input)
}

func (e *Executor) Pause() {
	e.pause.Open()
}

func (e *Executor) Resume() {
	e.pause.Close()
}

func (e *Executor) IsPaused() bool {
	return e.pause.IsClosed()
}

func (e *Executor) GetCycles() int64 {
	return atomic.LoadInt64(&e.stats.Cycles)
}

func (e *Executor) GetEnv() envPkg.IEnv {
	return e.env
}

type Configs struct {
	Env              envPkg.IEnv
	RateLimit        int
	RateLimitPeriod  time.Duration
	DefineImport     bool
	ImportCore       bool
	DoNotDeepCopyEnv bool
	DoNotProtectMaps bool
}

type VM struct {
	env              envPkg.IEnv
	rateLimit        int
	rateLimitPeriod  time.Duration
	importCore       bool
	defineImport     bool
	doNotDeepCopyEnv bool
	doNotProtectMaps bool
}

func New(configs *Configs) *VM {
	var env envPkg.IEnv
	var rateLimit int
	var rateLimitPeriod time.Duration
	var importCore bool
	var defineImport bool
	var doNotDeepCopyEnv bool
	var doNotProtectMaps bool
	if configs == nil || configs.Env == nil {
		env = envPkg.NewEnv()
	} else {
		env = configs.Env
	}
	if configs != nil {
		rateLimit = configs.RateLimit
		if configs.RateLimitPeriod == 0 {
			rateLimitPeriod = time.Second
		} else {
			rateLimitPeriod = configs.RateLimitPeriod
		}
		importCore = configs.ImportCore
		defineImport = configs.DefineImport
		doNotDeepCopyEnv = configs.DoNotDeepCopyEnv
		doNotProtectMaps = configs.DoNotProtectMaps
	}
	return &VM{
		env:              env,
		rateLimit:        rateLimit,
		rateLimitPeriod:  rateLimitPeriod,
		importCore:       importCore,
		defineImport:     defineImport,
		doNotDeepCopyEnv: doNotDeepCopyEnv,
		doNotProtectMaps: doNotProtectMaps,
	}
}

func (v *VM) Executor() IExecutor {
	return v.executor()
}

func (v *VM) Validate(ctx context.Context, val any) error {
	return v.validate(ctx, val)
}

func (v *VM) Has(ctx context.Context, val any, targets []any) ([]bool, error) {
	return v.has(ctx, val, targets)
}

func (v *VM) AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (envPkg.IEnv, error) {
	return v.env.AddPackage(name, methods, types)
}

func (v *VM) Define(k string, val any) error {
	return v.env.Define(k, val)
}

func (v *VM) DefineType(k string, val any) error {
	return v.env.DefineType(k, val)
}

func (v *VM) executor() *Executor {
	return NewExecutor(&ExecutorConfig{
		DoNotProtectMaps: v.doNotProtectMaps,
		DoNotDeepCopyEnv: v.doNotDeepCopyEnv,
		ImportCore:       v.importCore,
		DefineImport:     v.defineImport,
		RateLimit:        v.rateLimit,
		RateLimitPeriod:  v.rateLimitPeriod,
		Env:              v.env,
		WatchdogEnabled:  true,
		MaxEnvCount:      1000,
	})
}

func (v *VM) validate(ctx context.Context, val any) error {
	ctx = defaultCtx(ctx)
	e := v.executor()
	switch vv := val.(type) {
	case string:
		return e.validateWithContext(ctx, vv)
	case []byte:
		return e.validateCompiledWithContext(ctx, vv)
	default:
		return ErrInvalidInput
	}
}

func (v *VM) has(ctx context.Context, val any, targets []any) ([]bool, error) {
	ctx = defaultCtx(ctx)
	e := v.executor()
	switch vv := val.(type) {
	case string:
		return e.hasWithContext(ctx, vv, targets)
	case []byte:
		return e.hasCompiledWithContext(ctx, vv, targets)
	default:
		return nil, ErrInvalidInput
	}
}

type Result struct {
	Value reflect.Value
	Error error
}

type stats struct {
	Cycles int64
}

func incrCycle(vmp *vmParams) error {
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

func defaultCtx(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}

func srcToStmt(src string) (ast.Stmt, error) {
	return parser.ParseSrc(src)
}

func decode(by []byte) ast.Stmt {
	return compiler.Decode(by)
}

func (e *Executor) run(ctx context.Context, input any) (any, error) {
	ctx = defaultCtx(ctx)
	ctx, e.cancel = context.WithCancel(ctx)
	switch vv := input.(type) {
	case string:
		return e.executeWithContext(ctx, vv)
	case []byte:
		return e.executeCompiledWithContext(ctx, vv)
	case ast.Stmt:
		return e.runWithContext(ctx, vv)
	default:
		return nil, ErrInvalidInput
	}
}

func (e *Executor) runAsync(ctx context.Context, input any) {
	go func() {
		_, _ = e.run(ctx, input)
	}()
}

func (e *Executor) executeWithContext(ctx context.Context, src string) (any, error) {
	stmt, err := srcToStmt(src)
	if err != nil {
		return nilValue, err
	}
	return e.runWithContext(ctx, stmt)
}

func (e *Executor) validateWithContext(ctx context.Context, src string) error {
	stmt, err := srcToStmt(src)
	if err != nil {
		return err
	}
	return e.mainRunValidate(ctx, stmt)
}

func (e *Executor) hasWithContext(ctx context.Context, src string, targets []any) ([]bool, error) {
	stmt, err := srcToStmt(src)
	if err != nil {
		return nil, err
	}
	return e.hasAST(ctx, stmt, targets)
}

func (e *Executor) executeCompiledWithContext(ctx context.Context, src []byte) (any, error) {
	return e.runWithContext(ctx, decode(src))
}

func (e *Executor) validateCompiledWithContext(ctx context.Context, src []byte) error {
	return e.mainRunValidate(ctx, decode(src))
}

func (e *Executor) hasCompiledWithContext(ctx context.Context, src []byte, targets []any) ([]bool, error) {
	return e.hasAST(ctx, decode(src), targets)
}

func (e *Executor) runWithContext(ctx context.Context, stmts ast.Stmt) (any, error) {
	return valueToAny(e.mainRunNoTargets(ctx, stmts, false))
}

func (e *Executor) runWithContextForLoad(ctx context.Context, stmts ast.Stmt) (any, error) {
	return valueToAny(e.mainRunForLoad(ctx, stmts))
}

func valueToAny(rv reflect.Value, err error) (any, error) {
	if errors.Is(err, ErrReturn) {
		err = nil
	}
	if !rv.IsValid() || !rv.CanInterface() {
		return nil, err
	}
	return rv.Interface(), err
}

func (e *Executor) hasAST(ctx context.Context, stmt ast.Stmt, targets []any) (oks []bool, err error) {
	oks, _, err = e.mainRunWithWatchdog(ctx, stmt, true, targets)
	return
}

// we need to lock map operations MapIndex/SetMapIndex/mapIter.Next
// in the VM to avoid application crash if the script uses a map in a concurrent manner.
type mapLocker struct{ sync.Mutex }

func (m *mapLocker) Lock()   { m.Mutex.Lock() }
func (m *mapLocker) Unlock() { m.Mutex.Unlock() }

type vmParams struct {
	ctx              context.Context
	rvCh             chan Result
	stats            *stats
	doNotProtectMaps bool
	mapMutex         *mapLocker
	pause            *stateCh.StateCh
	rateLimit        *ratelimitanything.RateLimitAnything
	validate         bool
	has              map[any]bool
	validateLater    map[string]ast.Stmt
}

func newVmParams(ctx context.Context,
	rvCh chan Result,
	stats *stats,
	doNotProtectMaps bool,
	mapMutex *mapLocker,
	pause *stateCh.StateCh,
	rateLimit *ratelimitanything.RateLimitAnything,
	validate bool,
	has map[any]bool,
	validateLater map[string]ast.Stmt,
) *vmParams {
	return &vmParams{
		ctx:              ctx,
		rvCh:             rvCh,
		stats:            stats,
		doNotProtectMaps: doNotProtectMaps,
		mapMutex:         mapMutex,
		pause:            pause,
		rateLimit:        rateLimit,
		validate:         validate,
		has:              has,
		validateLater:    validateLater,
	}
}

func (e *Executor) mainRunValidate(ctx context.Context, stmt ast.Stmt) error {
	_, err := e.mainRunNoTargets(ctx, stmt, true)
	return err
}

// mainRunNoTargets executes statements in the specified environment.
func (e *Executor) mainRunNoTargets(ctx context.Context, stmt ast.Stmt, validate bool) (reflect.Value, error) {
	_, rv, err := e.mainRunWithWatchdog(ctx, stmt, validate, nil)
	return rv, err
}

func (e *Executor) mainRunForLoad(ctx context.Context, stmt ast.Stmt) (reflect.Value, error) {
	_, rv, err := e.mainRun(ctx, stmt, false, nil)
	return rv, err
}

func (e *Executor) watchdog(ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return
		}
		//fmt.Println(e.env.ChildCount())
		if e.env.ChildCount() > e.maxEnvCount.Load() {
			cancel()
			fmt.Println("KILLED")
			break
		}
	}
}

func (e *Executor) mainRunWithWatchdog(ctx context.Context, stmt ast.Stmt, validate bool, targets []any) ([]bool, reflect.Value, error) {
	if e.importCore {
		_ = e.env.Define("load", loadFn(e, ctx, validate))
	}

	// Start thread to watch for memory leaking scripts
	if e.watchdogEnabled {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		go e.watchdog(ctx, cancel)
	}

	return e.mainRun(ctx, stmt, validate, targets)
}

func (e *Executor) mainRun(ctx context.Context, stmt ast.Stmt, validate bool, targets []any) ([]bool, reflect.Value, error) {
	// We use rvCh because the script can start goroutines and crash in one of them.
	// So we need a way to stop the vm from another thread...

	stmt1, ok := stmt.(*ast.StmtsStmt)
	if !ok || stmt1 == nil {
		return nil, nilValue, ErrInvalidInput
	}

	envCopy := e.env

	runSingleStmtL := runSingleStmt

	oks := make([]bool, len(targets))
	has := make(map[any]bool)
	validateLater := make(map[string]ast.Stmt)
	for _, vv := range targets {
		has[fmt.Sprintf("%v", vv)] = false
	}
	rvCh := make(chan Result)
	vmp := newVmParams(ctx, rvCh, e.stats, e.doNotProtectMaps, e.mapMutex, e.pause, e.rateLimit, validate, has, validateLater)

	go func() {
		rv, err := runSingleStmtL(vmp, envCopy, stmt1)
		rvCh <- Result{Value: rv, Error: err}
	}()

	var result Result
	select {
	case result = <-rvCh:
	}

	if vmp.validate {
		for _, s := range vmp.validateLater {
			var err error
			envCopy.WithNewEnv(func(newenv envPkg.IEnv) {
				_, err = runSingleStmtL(vmp, newenv, s)
			})
			if err != nil {
				return nil, nilValue, err
			}
		}
		for i, vv := range targets {
			oks[i] = has[fmt.Sprintf("%v", vv)]
		}
	}

	return oks, result.Value, result.Error
}
