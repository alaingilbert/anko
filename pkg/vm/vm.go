package vm

import (
	"context"
	"github.com/alaingilbert/anko/pkg/packages"
	"github.com/alaingilbert/anko/pkg/utils"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/executor"
	"time"
)

type IVM interface {
	Executor() executor.IExecutor
	Validate(context.Context, any) error
	Has(context.Context, any, []any) ([]bool, error)

	Define(k string, v any) error
	DefineType(k string, v any) error
	AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (envPkg.IEnv, error)
}

// Compile time checks to ensure type satisfies IVM interface
var _ IVM = (*VM)(nil)

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

func (v *VM) Executor() executor.IExecutor {
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

func (v *VM) executor() *executor.Executor {
	return executor.NewExecutor(&executor.Config{
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
	ctx = utils.DefaultCtx(ctx)
	e := v.executor()
	switch vv := val.(type) {
	case string:
		return e.ValidateWithContext(ctx, vv)
	case []byte:
		return e.ValidateCompiledWithContext(ctx, vv)
	default:
		return executor.ErrInvalidInput
	}
}

func (v *VM) has(ctx context.Context, val any, targets []any) ([]bool, error) {
	ctx = utils.DefaultCtx(ctx)
	e := v.executor()
	switch vv := val.(type) {
	case string:
		return e.HasWithContext(ctx, vv, targets)
	case []byte:
		return e.HasCompiledWithContext(ctx, vv, targets)
	default:
		return nil, executor.ErrInvalidInput
	}
}
