package vm

import (
	"context"
	"github.com/alaingilbert/anko/pkg/packages"
	"github.com/alaingilbert/anko/pkg/utils"
	envPkg "github.com/alaingilbert/anko/pkg/vm/env"
	"github.com/alaingilbert/anko/pkg/vm/executor"
	"time"
)

// IVM interface that VM implements
type IVM interface {
	Executor(*executor.Config) executor.IExecutor
	Validate(context.Context, any) error
	Has(context.Context, any, []any) ([]bool, error)

	Define(k string, v any) error
	DefineCtx(k string, v any) error
	DefineType(k string, v any) error
	AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (envPkg.IEnv, error)
}

// Compile time checks to ensure type satisfies IVM interface
var _ IVM = (*VM)(nil)

// Config for the vm
type Config struct {
	Env             envPkg.IEnv
	RateLimit       *int
	RateLimitPeriod *time.Duration
	DefineImport    *bool
	ImportCore      *bool
	DeepCopyEnv     *bool
	ProtectMaps     *bool
	DbgEnabled      *bool
	Watchdog        *bool
	MaxEnvCount     *int
	ResetEnv        *bool
}

// VM base vm
type VM struct {
	env             envPkg.IEnv
	rateLimit       *int
	rateLimitPeriod *time.Duration
	importCore      *bool
	defineImport    *bool
	deepCopyEnv     *bool
	protectMaps     *bool
	dbgEnabled      *bool
	watchdog        *bool
	maxEnvCount     *int
	resetEnv        *bool
}

// New creates a new vm
func New(config *Config) *VM {
	var env envPkg.IEnv
	if config == nil || config.Env == nil {
		env = envPkg.NewEnv()
	} else {
		env = config.Env
	}
	v := &VM{env: env}
	if config != nil {
		v.rateLimit = config.RateLimit
		v.rateLimitPeriod = config.RateLimitPeriod
		v.importCore = config.ImportCore
		v.defineImport = config.DefineImport
		v.deepCopyEnv = config.DeepCopyEnv
		v.protectMaps = config.ProtectMaps
		v.dbgEnabled = config.DbgEnabled
		v.watchdog = config.Watchdog
		v.maxEnvCount = config.MaxEnvCount
		v.resetEnv = config.ResetEnv
	}
	return v
}

// Executor creates a new executor
func (v *VM) Executor(cfg *executor.Config) executor.IExecutor {
	return v.executor(cfg)
}

// Validate a script without executing it
func (v *VM) Validate(ctx context.Context, val any) error {
	return v.validate(ctx, val)
}

// Has returns either or not the targets are being used in the script, without executing the script
func (v *VM) Has(ctx context.Context, val any, targets []any) ([]bool, error) {
	return v.has(ctx, val, targets)
}

// AddPackage adds a package in the Env
func (v *VM) AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (envPkg.IEnv, error) {
	return v.env.AddPackage(name, methods, types)
}

// Define defines a key/value in the Env
func (v *VM) Define(k string, val any) error {
	return v.env.Define(k, val)
}

// DefineCtx defines a key/value in the Env. If val is a function,
// the running context will be injected in its arguments
func (v *VM) DefineCtx(k string, val any) error {
	return v.env.DefineCtx(k, val)
}

// DefineType defines a new type in the Env
func (v *VM) DefineType(k string, val any) error {
	return v.env.DefineType(k, val)
}

func (v *VM) getDefaultExecutorConfig() *executor.Config {
	return &executor.Config{
		ProtectMaps:     v.protectMaps,
		DeepCopyEnv:     v.deepCopyEnv,
		ImportCore:      v.importCore,
		DefineImport:    v.defineImport,
		RateLimit:       v.rateLimit,
		RateLimitPeriod: v.rateLimitPeriod,
		Env:             v.env,
		DbgEnabled:      v.dbgEnabled,
		Watchdog:        v.watchdog,
		MaxEnvCount:     v.maxEnvCount,
		ResetEnv:        v.resetEnv,
	}
}

func (v *VM) executor(cfg *executor.Config) *executor.Executor {
	cfgToUse := v.getDefaultExecutorConfig()
	if cfg != nil {
		if cfg.Env != nil {
			cfgToUse.Env = cfg.Env
		}
		cfgToUse.RateLimit = utils.Override(cfgToUse.RateLimit, cfg.RateLimit)
		cfgToUse.RateLimitPeriod = utils.Override(cfgToUse.RateLimitPeriod, cfg.RateLimitPeriod)
		cfgToUse.Watchdog = utils.Override(cfgToUse.Watchdog, cfg.Watchdog)
		cfgToUse.MaxEnvCount = utils.Override(cfgToUse.MaxEnvCount, cfg.MaxEnvCount)
		cfgToUse.ProtectMaps = utils.Override(cfgToUse.ProtectMaps, cfg.ProtectMaps)
		cfgToUse.DeepCopyEnv = utils.Override(cfgToUse.DeepCopyEnv, cfg.DeepCopyEnv)
		cfgToUse.ImportCore = utils.Override(cfgToUse.ImportCore, cfg.ImportCore)
		cfgToUse.DefineImport = utils.Override(cfgToUse.DefineImport, cfg.DefineImport)
		cfgToUse.DbgEnabled = utils.Override(cfgToUse.DbgEnabled, cfg.DbgEnabled)
		cfgToUse.ResetEnv = utils.Override(cfgToUse.ResetEnv, cfg.ResetEnv)
	}
	return executor.NewExecutor(cfgToUse)
}

func (v *VM) validate(ctx context.Context, val any) error {
	return v.executor(nil).Validate(utils.DefaultCtx(ctx), val)
}

func (v *VM) has(ctx context.Context, val any, targets []any) ([]bool, error) {
	return v.executor(nil).Has(utils.DefaultCtx(ctx), val, targets)
}
