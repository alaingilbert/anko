package vm

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/compiler"
	"github.com/alaingilbert/anko/pkg/parser"
	"github.com/alaingilbert/anko/pkg/utils"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
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
	AddPackage(name string, methods map[string]any, types map[string]any) (*envPkg.Env, error)

	GetEnv() *envPkg.Env
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

	GetEnv() *envPkg.Env
}

var _ IExecutor = (*Executor)(nil)

type Executor struct {
	env       *envPkg.Env
	pause     *StateCh
	stats     *stats
	rateLimit *utils.RateLimitAnything
	mapMutex  *sync.Mutex
	cancel    context.CancelFunc
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

func (e *Executor) GetEnv() *envPkg.Env {
	return e.env
}

type Configs struct {
	Env              *envPkg.Env
	RateLimit        int
	RateLimitPeriod  time.Duration
	DefineImport     bool
	ImportCore       bool
	DoNotDeepCopyEnv bool
}

type VM struct {
	env              *envPkg.Env
	rateLimit        int
	rateLimitPeriod  time.Duration
	importCore       bool
	defineImport     bool
	doNotDeepCopyEnv bool
}

func New(configs *Configs) *VM {
	var env *envPkg.Env
	if configs == nil || configs.Env == nil {
		env = envPkg.NewEnv()
	} else {
		env = configs.Env
	}
	var rateLimit int
	if configs != nil {
		rateLimit = configs.RateLimit
	}
	var rateLimitPeriod time.Duration
	if configs != nil {
		if configs.RateLimitPeriod == 0 {
			rateLimitPeriod = time.Second
		} else {
			rateLimitPeriod = configs.RateLimitPeriod
		}
	}
	var importCore bool
	if configs != nil {
		importCore = configs.ImportCore
	}
	var defineImport bool
	if configs != nil {
		defineImport = configs.DefineImport
	}
	var doNotDeepCopyEnv bool
	if configs != nil {
		doNotDeepCopyEnv = configs.DoNotDeepCopyEnv
	}
	return &VM{
		env:              env,
		rateLimit:        rateLimit,
		rateLimitPeriod:  rateLimitPeriod,
		importCore:       importCore,
		defineImport:     defineImport,
		doNotDeepCopyEnv: doNotDeepCopyEnv,
	}
}

func (v *VM) Executor() IExecutor {
	return v.executor()
}

func (v *VM) executor() *Executor {
	e := &Executor{}
	if v.doNotDeepCopyEnv {
		e.env = v.env
	} else {
		e.env = v.env.DeepCopy()
	}
	if v.importCore {
		Import(e.env)
	}
	if v.defineImport {
		DefineImport(e.env)
	}
	e.pause = NewStateCh(true)
	e.stats = &stats{}
	e.mapMutex = &sync.Mutex{}
	if v.rateLimit > 0 {
		e.rateLimit = utils.NewRateLimitAnything(int64(v.rateLimit), v.rateLimitPeriod)
	}
	return e
}

func (v *VM) Validate(ctx context.Context, val any) error {
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

func (v *VM) Has(ctx context.Context, val any, targets []any) ([]bool, error) {
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

func (v *VM) AddPackage(name string, methods map[string]any, types map[string]any) (*envPkg.Env, error) {
	return v.env.AddPackage(name, methods, types)
}

func (v *VM) GetEnv() *envPkg.Env {
	return v.env
}

func (v *VM) Define(k string, val any) error {
	return v.env.Define(k, val)
}

func (v *VM) DefineType(k string, val any) error {
	return v.env.DefineType(k, val)
}

type Result struct {
	Value reflect.Value
	Error error
}

type stats struct {
	Cycles int64
}

func incrCycle(vmp *vmParams) error {
	select {
	case <-vmp.ctx.Done():
		return ErrInterrupt
	default:
	}
	select {
	case <-vmp.pause.Wait():
	case <-vmp.ctx.Done():
		return ErrInterrupt
	}
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
	rv, err := e.mainRun(ctx, stmts, false)
	if errors.Is(err, ErrReturn) {
		err = nil
	}
	if !rv.IsValid() || !rv.CanInterface() {
		return nil, err
	}
	return rv.Interface(), err
}

func (e *Executor) validate(ctx context.Context, src string) (any, error) {
	stmts, err := srcToStmt(src)
	if err != nil {
		return nil, err
	}
	rv, err := e.mainRun(ctx, stmts, true)
	if errors.Is(err, ErrReturn) {
		err = nil
	}
	if !rv.IsValid() || !rv.CanInterface() {
		return nil, err
	}
	return rv.Interface(), err
}

func (e *Executor) hasAST(ctx context.Context, stmt ast.Stmt, targets []any) (oks []bool, err error) {
	oks, _, err = e.mainRun1(ctx, stmt, true, targets)
	return
}

type vmParams struct {
	ctx           context.Context
	rvCh          chan Result
	stats         *stats
	mapMutex      *sync.Mutex
	pause         *StateCh
	rateLimit     *utils.RateLimitAnything
	validate      bool
	has           map[any]bool
	validateLater map[string]ast.Stmt
}

func newVmParams(ctx context.Context,
	rvCh chan Result,
	stats *stats,
	mapMutex *sync.Mutex,
	pause *StateCh,
	rateLimit *utils.RateLimitAnything,
	validate bool,
	has map[any]bool,
	validateLater map[string]ast.Stmt,
) *vmParams {
	return &vmParams{
		ctx:           ctx,
		rvCh:          rvCh,
		stats:         stats,
		mapMutex:      mapMutex,
		pause:         pause,
		rateLimit:     rateLimit,
		validate:      validate,
		has:           has,
		validateLater: validateLater,
	}
}

// mainRun executes statements in the specified environment.
func (e *Executor) mainRun(ctx context.Context, stmt ast.Stmt, validate bool) (reflect.Value, error) {
	_, rv, err := e.mainRun1(ctx, stmt, validate, nil)
	return rv, err
}

func (e *Executor) mainRunValidate(ctx context.Context, stmt ast.Stmt) error {
	_, _, err := e.mainRun1(ctx, stmt, true, nil)
	return err
}

func (e *Executor) mainRun1(ctx context.Context, stmt ast.Stmt, validate bool, targets []any) ([]bool, reflect.Value, error) {
	// We use rvCh because the script can start goroutines and crash in one of them.
	// So we need a way to stop the vm from another thread...
	oks := make([]bool, len(targets))
	has := make(map[any]bool)
	validateLater := make(map[string]ast.Stmt)
	for _, vv := range targets {
		has[fmt.Sprintf("%v", vv)] = false
	}
	envCopy := e.env
	rvCh := make(chan Result)
	vmp := newVmParams(ctx, rvCh, e.stats, e.mapMutex, e.pause, e.rateLimit, validate, has, validateLater)
	go func() {
		rv, err := runSingleStmt(vmp, envCopy, stmt)
		rvCh <- Result{Value: rv, Error: err}
	}()

	var result Result
	select {
	case result = <-rvCh:
	}

	if vmp.validate {
		for _, s := range vmp.validateLater {
			newEnv := envCopy.NewEnv()
			_, err := runSingleStmt(vmp, newEnv, s)
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
