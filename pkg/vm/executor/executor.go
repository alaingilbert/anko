package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	"github.com/alaingilbert/anko/pkg/utils/ratelimitanything"
	"github.com/alaingilbert/anko/pkg/utils/stateCh"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/runner"
	"github.com/alaingilbert/mtx"
	"os"
	"reflect"
	"sync/atomic"
	"time"
)

type IExecutor interface {
	DefineCtx(string, any) error
	GetCycles() int64
	Has(ctx context.Context, input any, targets []any) ([]bool, error)
	IsPaused() bool
	IsRunning() bool
	Pause()
	Resume()
	Run(ctx context.Context, input any) (any, error)
	RunAsync(ctx context.Context, input any)
	Stop()
	Validate(ctx context.Context, input any) error

	GetEnv() envPkg.IEnv
}

var _ IExecutor = (*Executor)(nil)

type Executor struct {
	env              envPkg.IEnv                          // executor's env
	pause            *stateCh.StateCh                     // allows pause/resume of scripts
	stats            *runner.Stats                        // keep track of stmt/expr processed
	rateLimit        *ratelimitanything.RateLimitAnything // rate limit expr processed/duration
	doNotProtectMaps bool                                 // either or not to protect maps operations in the VM
	mapMutex         *runner.MapLocker                    // locker object to protect maps
	cancel           context.CancelFunc                   // use to Stop a script
	importCore       bool                                 // either or not to import core functions in executor's env
	watchdogEnabled  bool                                 // either or not to run the watchdog
	maxEnvCount      *mtx.Mtx[int64]                      // maximum sub-env allowed before the watchdog kills the script
	isRunning        atomic.Bool
}

type Config struct {
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

func NewExecutor(cfg *Config) *Executor {
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
		runner.Import(e.env)
	}
	if cfg.DefineImport {
		runner.DefineImport(e.env)
	}
	if cfg.WatchdogEnabled {
		e.watchdogEnabled = cfg.WatchdogEnabled
	}
	e.pause = stateCh.NewStateCh(true)
	e.stats = &runner.Stats{}
	e.doNotProtectMaps = cfg.DoNotProtectMaps
	e.importCore = cfg.ImportCore
	e.mapMutex = &runner.MapLocker{}
	e.maxEnvCount = mtx.NewRWMtxPtr(int64(cfg.MaxEnvCount))
	if cfg.RateLimit > 0 {
		e.rateLimit = ratelimitanything.NewRateLimitAnything(int64(cfg.RateLimit), cfg.RateLimitPeriod)
	}
	return e
}

func (e *Executor) Stop() {
	e.stop()
}

func (e *Executor) Run(ctx context.Context, input any) (any, error) {
	return e.run(ctx, input)
}

func (e *Executor) RunAsync(ctx context.Context, input any) {
	e.runAsync(ctx, input)
}

func (e *Executor) Validate(ctx context.Context, input any) error {
	return e.validate(ctx, input)
}

func (e *Executor) Has(ctx context.Context, input any, targets []any) ([]bool, error) {
	return e.has(ctx, input, targets)
}

func (e *Executor) Pause() {
	e.pauseFn()
}

func (e *Executor) Resume() {
	e.resume()
}

func (e *Executor) IsPaused() bool {
	return !e.pause.IsClosed()
}

func (e *Executor) IsRunning() bool {
	return e.isRunning.Load()
}

func (e *Executor) GetCycles() int64 {
	return atomic.LoadInt64(&e.stats.Cycles)
}

func (e *Executor) DefineCtx(key string, value any) error {
	return e.defineCtx(key, value)
}

func (e *Executor) GetEnv() envPkg.IEnv {
	return e.env
}

func (e *Executor) run(ctx context.Context, input any) (any, error) {
	if !e.isRunning.CompareAndSwap(false, true) {
		return nil, ErrAlreadyRunning
	}
	defer e.isRunning.Store(false)
	ctx = utils.DefaultCtx(ctx)
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

func (e *Executor) validate(ctx context.Context, input any) error {
	switch vv := input.(type) {
	case string:
		return e.ValidateWithContext(ctx, vv)
	case []byte:
		return e.ValidateCompiledWithContext(ctx, vv)
	default:
		return ErrInvalidInput
	}
}

func (e *Executor) has(ctx context.Context, input any, targets []any) ([]bool, error) {
	switch vv := input.(type) {
	case string:
		return e.HasWithContext(ctx, vv, targets)
	case []byte:
		return e.HasCompiledWithContext(ctx, vv, targets)
	default:
		return nil, ErrInvalidInput
	}
}

func (e *Executor) runAsync(ctx context.Context, input any) {
	go func() {
		if rv, err := e.run(ctx, input); err != nil {
			fmt.Println(rv, err)
		}
	}()
}

func (e *Executor) stop() {
	if e.cancel != nil {
		e.cancel()
	}
	e.resume()
}

func (e *Executor) pauseFn() {
	if e.isRunning.Load() {
		e.pause.Open()
	}
}

func (e *Executor) resume() {
	e.pause.Close()
}

func (e *Executor) defineCtx(key string, value any) error {
	return e.env.DefineCtx(key, value)
}

func srcToStmt(src string) (ast.Stmt, error) {
	return parser.ParseSrc(src)
}

func decode(by []byte) ast.Stmt {
	return compiler.Decode(by)
}

func (e *Executor) executeWithContext(ctx context.Context, src string) (any, error) {
	stmt, err := srcToStmt(src)
	if err != nil {
		return nilValue, err
	}
	return e.runWithContext(ctx, stmt)
}

func (e *Executor) ValidateWithContext(ctx context.Context, src string) error {
	stmt, err := srcToStmt(src)
	if err != nil {
		return err
	}
	return e.mainRunValidate(ctx, stmt)
}

func (e *Executor) HasWithContext(ctx context.Context, src string, targets []any) ([]bool, error) {
	stmt, err := srcToStmt(src)
	if err != nil {
		return nil, err
	}
	return e.hasAST(ctx, stmt, targets)
}

func (e *Executor) executeCompiledWithContext(ctx context.Context, src []byte) (any, error) {
	return e.runWithContext(ctx, decode(src))
}

func (e *Executor) ValidateCompiledWithContext(ctx context.Context, src []byte) error {
	return e.mainRunValidate(ctx, decode(src))
}

func (e *Executor) HasCompiledWithContext(ctx context.Context, src []byte, targets []any) ([]bool, error) {
	return e.hasAST(ctx, decode(src), targets)
}

func (e *Executor) runWithContext(ctx context.Context, stmts ast.Stmt) (any, error) {
	return valueToAny(e.mainRunNoTargets(ctx, stmts, false))
}

func (e *Executor) runWithContextForLoad(ctx context.Context, stmts ast.Stmt) (any, error) {
	return valueToAny(e.mainRunForLoad(ctx, stmts))
}

func valueToAny(rv reflect.Value, err error) (any, error) {
	if errors.Is(err, runner.ErrReturn) {
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

// Dynamically load a file and execute it, return the RV value
func (e *Executor) loadFn(ctx context.Context, validate bool) func(string) any {
	return func(s string) any {
		if validate {
			return nilValue
		}
		body, err := os.ReadFile(s)
		if err != nil {
			panic(err)
		}
		scanner := new(parser.Scanner)
		scanner.Init(string(body))
		stmts, err := parser.Parse(scanner)
		if err != nil {
			var pe *parser.Error
			if errors.As(err, &pe) {
				pe.Filename = s
				panic(pe)
			}
			panic(err)
		}
		rv, err := e.runWithContextForLoad(ctx, stmts)
		if err != nil {
			panic(err)
		}
		return rv
	}
}

func (e *Executor) mainRunWithWatchdog(ctx context.Context, stmt ast.Stmt, validate bool, targets []any) ([]bool, reflect.Value, error) {
	if e.importCore {
		_ = e.env.Define("load", e.loadFn(ctx, validate))
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

var nilValue = reflect.New(reflect.TypeOf((*any)(nil)).Elem()).Elem()
var ErrInvalidInput = errors.New("invalid input")
var ErrAlreadyRunning = errors.New("executor already running")

func (e *Executor) mainRun(ctx context.Context, stmt ast.Stmt, validate bool, targets []any) ([]bool, reflect.Value, error) {
	stmt1, ok := stmt.(*ast.StmtsStmt)
	if !ok || stmt1 == nil {
		return nil, nilValue, ErrInvalidInput
	}

	envCopy := e.env

	has := make(map[any]bool)
	for _, vv := range targets {
		has[fmt.Sprintf("%v", vv)] = false
	}

	rv, err := runner.Run(&runner.Config{
		Ctx:              ctx,
		Env:              envCopy,
		Stmt:             stmt1,
		Stats:            e.stats,
		DoNotProtectMaps: e.doNotProtectMaps,
		MapMutex:         e.mapMutex,
		Pause:            e.pause,
		RateLimit:        e.rateLimit,
		Validate:         validate,
		Has:              has,
	})
	if err != nil {
		return nil, rv, err
	}

	oks := make([]bool, len(targets))
	if validate {
		for i, vv := range targets {
			oks[i] = has[fmt.Sprintf("%v", vv)]
		}
	}

	return oks, rv, err
}
