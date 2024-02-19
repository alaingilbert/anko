package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	"github.com/alaingilbert/anko/pkg/utils/pubsub"
	"github.com/alaingilbert/anko/pkg/utils/ratelimitanything"
	"github.com/alaingilbert/anko/pkg/utils/stateCh"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/runner"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"github.com/alaingilbert/mtx"
	"os"
	"reflect"
	"sync/atomic"
	"time"
)

// Sub subscriber type for executor events
type Sub = pubsub.Sub[string, Evt]

// IExecutor interface that the executor implements
type IExecutor interface {
	GetRateLimit() (int64, time.Duration)
	Has(ctx context.Context, input any, targets []any) ([]bool, error)
	IsPaused() bool
	IsRunning() bool
	Pause() bool
	Resume() bool
	Run(ctx context.Context, input any) (any, error)
	RunAsync(ctx context.Context, input any) bool
	SetRateLimit(int64, time.Duration)
	Stop() bool
	Subscribe() *Sub
	TogglePause() TogglePauseResult
	Validate(ctx context.Context, input any) error

	GetEnv() envPkg.IEnv
}

// Compile time checks to ensure type satisfies IExecutor interface
var _ IExecutor = (*Executor)(nil)

const (
	executorTopic = "executor"
)

// Evt events spawned by the executor
type Evt int

func (e Evt) String() (out string) {
	switch e {
	case StartedEvt:
		out = "started"
	case CompletedEvt:
		out = "completed"
	case PausedEvt:
		out = "paused"
	case ResumedEvt:
		out = "resumed"
	}
	return
}

const (
	StartedEvt Evt = iota + 1
	CompletedEvt
	PausedEvt
	ResumedEvt
)

// Executor is responsible for executing scripts and managing state
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
	isRunning        atomic.Bool                          // either or not the executor is running a script
	pubSubEvts       *pubsub.PubSub[string, Evt]          // pubsub for executor's events
	dbgEnabled       bool                                 // either or not to enable dbg()
}

// Config for the executor
type Config struct {
	ProtectMaps     *bool
	DeepCopyEnv     *bool
	ImportCore      *bool
	Watchdog        *bool
	DefineImport    *bool
	DbgEnabled      *bool
	RateLimit       int
	RateLimitPeriod time.Duration
	Env             envPkg.IEnv
	MaxEnvCount     *int
}

// NewExecutor creates a new executor
func NewExecutor(cfg *Config) *Executor {
	if cfg == nil {
		return nil
	}
	e := &Executor{}
	deepCopyEnv := utils.Default(cfg.DeepCopyEnv, true)
	importCore := utils.Default(cfg.ImportCore, false)
	defineImport := utils.Default(cfg.DefineImport, false)
	if deepCopyEnv {
		e.env = cfg.Env.DeepCopy()
	} else {
		e.env = cfg.Env
	}
	if importCore {
		runner.Import(e.env)
	}
	if defineImport {
		runner.DefineImport(e.env)
	}
	e.pause = stateCh.NewStateCh(true)
	e.stats = &runner.Stats{}
	e.importCore = utils.Default(cfg.ImportCore, false)
	e.dbgEnabled = utils.Default(cfg.DbgEnabled, true)
	e.doNotProtectMaps = utils.Default(cfg.ProtectMaps, true)
	e.mapMutex = &runner.MapLocker{}
	e.watchdogEnabled = utils.Default(cfg.Watchdog, true)
	e.maxEnvCount = mtx.NewRWMtxPtr(int64(utils.Default(cfg.MaxEnvCount, 1000)))
	e.rateLimit = ratelimitanything.NewRateLimitAnything(int64(cfg.RateLimit), cfg.RateLimitPeriod)
	e.pubSubEvts = pubsub.NewPubSub[Evt](nil)
	return e
}

// Stop the execution of a script. Returns true if we stopped it, false if the executor was not running anything.
func (e *Executor) Stop() bool {
	return e.stop()
}

// Run the input synchronously
func (e *Executor) Run(ctx context.Context, input any) (any, error) {
	return e.run(ctx, input)
}

// RunAsync returns true if the script is being run async, false if we did not start it
func (e *Executor) RunAsync(ctx context.Context, input any) bool {
	return e.runAsync(ctx, input)
}

// Validate the input. It does not actually run the script, only statically analyze the input.
func (e *Executor) Validate(ctx context.Context, input any) error {
	return e.validate(ctx, input)
}

// Has receives targets, and return either or not these targets are being used in the script.
// It does not actually run the script, it only statically check if the targets are used or not.
func (e *Executor) Has(ctx context.Context, input any, targets []any) ([]bool, error) {
	return e.has(ctx, input, targets)
}

// TogglePause toggle the pause state. Returns NoopToggle if the pause state did not change.
func (e *Executor) TogglePause() TogglePauseResult {
	return e.togglePause()
}

// Pause the execution of the script. Return true if the script got paused, false if it was already paused.
func (e *Executor) Pause() bool {
	return e.pauseFn()
}

// GetRateLimit returns the current set rate limit
func (e *Executor) GetRateLimit() (int64, time.Duration) {
	return e.getRateLimit()
}

// SetRateLimit set rate limit
func (e *Executor) SetRateLimit(limit int64, period time.Duration) {
	e.setRateLimit(limit, period)
}

// Resume the execution of the script. Return true if the script got resumed, false if it was already running.
func (e *Executor) Resume() bool {
	return e.resume()
}

// Subscribe returns a subscriber to the executor events
func (e *Executor) Subscribe() *Sub {
	return e.pubSubEvts.Subscribe(executorTopic)
}

// IsPaused returns either or not the execution is paused
func (e *Executor) IsPaused() bool {
	return !e.pause.IsClosed()
}

// IsRunning returns either or not the executor is currently running a script
func (e *Executor) IsRunning() bool {
	return e.isRunning.Load()
}

// GetEnv returns the Env used by the executor
func (e *Executor) GetEnv() envPkg.IEnv {
	return e.env
}

func (e *Executor) run(ctx context.Context, input any) (any, error) {
	if !e.isRunning.CompareAndSwap(false, true) {
		return nil, ErrAlreadyRunning
	}
	defer e.isRunning.Store(false)
	e.pubSubEvts.Pub(executorTopic, StartedEvt)
	defer e.pubSubEvts.Pub(executorTopic, CompletedEvt)
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

func (e *Executor) runAsync(ctx context.Context, input any) bool {
	if e.isRunning.Load() {
		return false
	}
	go func() {
		if rv, err := e.run(ctx, input); err != nil {
			fmt.Println(rv, err)
		}
	}()
	return true
}

func (e *Executor) stop() bool {
	if !e.isRunning.Load() {
		return false
	}
	if e.cancel != nil {
		e.cancel()
	}
	e.resume()
	return true
}

type TogglePauseResult int

const (
	NoopToggle TogglePauseResult = iota
	PausedToggle
	ResumedToggle
)

func (e *Executor) togglePause() TogglePauseResult {
	if e.isRunning.Load() {
		if e.pause.Toggle() {
			e.pubSubEvts.Pub(executorTopic, PausedEvt)
			return PausedToggle
		}
		e.pubSubEvts.Pub(executorTopic, ResumedEvt)
		return ResumedToggle
	}
	return NoopToggle
}

func (e *Executor) pauseFn() (changed bool) {
	if e.isRunning.Load() {
		changed = e.pause.Open()
		if changed {
			e.pubSubEvts.Pub(executorTopic, PausedEvt)
		}
	}
	return changed
}

func (e *Executor) getRateLimit() (int64, time.Duration) {
	return e.rateLimit.GetLimit()
}

func (e *Executor) setRateLimit(limit int64, period time.Duration) {
	e.rateLimit.Set(limit, period)
}

func (e *Executor) resume() (changed bool) {
	changed = e.pause.Close()
	if changed {
		e.pubSubEvts.Pub(executorTopic, ResumedEvt)
	}
	return changed
}

// getCycles returns how many expr/stmt were processed by the executor
func (e *Executor) getCycles() int64 {
	return atomic.LoadInt64(&e.stats.Cycles)
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

var nilValue = vmUtils.NilValue
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
		Ctx:         ctx,
		Env:         envCopy,
		Stmt:        stmt1,
		Stats:       e.stats,
		ProtectMaps: e.doNotProtectMaps,
		MapMutex:    e.mapMutex,
		Pause:       e.pause,
		RateLimit:   e.rateLimit,
		DbgEnabled:  e.dbgEnabled,
		Validate:    validate,
		Has:         has,
	})
	if errors.Is(err, runner.ErrReturn) {
		err = nil
	}
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
