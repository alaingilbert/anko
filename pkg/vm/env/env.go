package env

import (
	"bytes"
	"fmt"
	"github.com/alaingilbert/anko/pkg/packages"
	"github.com/alaingilbert/anko/pkg/utils"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"github.com/alaingilbert/mtx"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
)

// It's fine for this map to not be mutex protected, so long as no one writes in it
var basicTypes = map[string]reflect.Type{
	"interface": reflect.ValueOf([]any{int64(1)}).Index(0).Type(),
	"any":       reflect.ValueOf([]any{int64(1)}).Index(0).Type(),
	"bool":      reflect.TypeOf(true),
	"string":    reflect.TypeOf("a"),
	"int":       reflect.TypeOf(int(1)),
	"int8":      reflect.TypeOf(int8(1)),
	"int16":     reflect.TypeOf(int16(1)),
	"int32":     reflect.TypeOf(int32(1)),
	"int64":     reflect.TypeOf(int64(1)),
	"uint":      reflect.TypeOf(uint(1)),
	"uint8":     reflect.TypeOf(uint8(1)),
	"uint16":    reflect.TypeOf(uint16(1)),
	"uint32":    reflect.TypeOf(uint32(1)),
	"uint64":    reflect.TypeOf(uint64(1)),
	"uintptr":   reflect.TypeOf(uintptr(1)),
	"byte":      reflect.TypeOf(byte(1)),
	"rune":      reflect.TypeOf('a'),
	"float32":   reflect.TypeOf(float32(1)),
	"float64":   reflect.TypeOf(float64(1)),
	"error":     reflect.TypeOf((*error)(nil)).Elem(),
}

// CapturedFunc is stacked in the scope
type CapturedFunc struct {
	Func      reflect.Value
	Args      []reflect.Value
	CallSlice bool
}

type IEnv interface {
	AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (IEnv, error)
	ChildCount() int64
	DeepCopy() IEnv
	Defers() *mtx.Slice[CapturedFunc]
	Define(k string, v any) error
	DefineCtx(k string, v any) error
	DefineGlobalValue(k string, v reflect.Value) error
	DefineReflectType(k string, t reflect.Type) error
	DefineType(k string, t any) error
	DefineValue(k string, v reflect.Value) error
	Delete(k string) error
	DeleteGlobal(k string) error
	Destroy()
	Get(k string) (any, error)
	GetEnvFromPath(path []string) (IEnv, error)
	GetValue(k string) (reflect.Value, error)
	Name() string
	NewEnv() IEnv
	WithNewEnv(func(IEnv))
	NewModule(symbol string) (IEnv, error)
	SetValue(k string, v reflect.Value) error
	String() string
	Type(k string) (reflect.Type, error)
	Types() *mtx.Map[string, reflect.Type]
	Values() *mtx.Map[string, reflect.Value]
}

var _ IEnv = (*Env)(nil)

// Env provides interface to run VM. This mean function scope and blocked-scope.
// If stack goes to blocked-scope, it will make new Env.
type Env struct {
	childCountVar int64 // atomic
	parent        *Env
	name          *mtx.Mtx[string]
	values        *mtx.Map[string, reflect.Value]
	types         *mtx.Map[string, reflect.Type]
	defers        *mtx.Slice[CapturedFunc]
}

// NewEnv creates new global scope.
func NewEnv() *Env { return newEnv() }

// NewEnv creates new child scope.
func (e *Env) NewEnv() IEnv { return e.newEnv() }

func (e *Env) WithNewEnv(clb func(IEnv)) { e.withNewEnv(clb) }

// NewModule creates new child scope and define it as a symbol.
// This is a shortcut for calling e.NewEnv then Define that new Env.
func (e *Env) NewModule(symbol string) (IEnv, error) { return e.newModule(symbol) }

func (e *Env) Destroy() { e.destroy() }

func (e *Env) ChildCount() int64 { return e.childCount() }

func (e *Env) Values() *mtx.Map[string, reflect.Value] { return e.values }

func (e *Env) Types() *mtx.Map[string, reflect.Type] { return e.types }

func (e *Env) Defers() *mtx.Slice[CapturedFunc] { return e.defers }

// GetEnvFromPath returns Env from path
func (e *Env) GetEnvFromPath(path []string) (IEnv, error) { return e.getEnvFromPath(path) }

// AddPackage creates a new env with a name that has methods and types in it. Created under the parent env
func (e *Env) AddPackage(name string, methods packages.PackageMap, types packages.PackageMap) (IEnv, error) {
	return e.addPackage(name, methods, types)
}

func (e *Env) Name() string { return e.getName() }

// String returns string of values and types in current scope.
func (e *Env) String() string { return e.string() }

// Addr returns pointer value which specified symbol. It goes to upper scope until
// found or returns error.
//func (e *Env) Addr(k string) (reflect.Value, error) { return e.addr(k) }

// Type returns type which specified symbol. It goes to upper scope until
// found or returns error.
func (e *Env) Type(k string) (reflect.Type, error) { return e.typ(k) }

// Get returns value which specified symbol. It goes to upper scope until
// found or returns error.
func (e *Env) Get(k string) (any, error) { return e.get(k) }

func (e *Env) GetValue(k string) (reflect.Value, error) { return e.getValue(k) }

//func (e *Env) Set(k string, v any) error { return e.set(k, v) }

func (e *Env) SetValue(k string, v reflect.Value) error { return e.setValue(k, v) }

// DefineGlobal defines symbol in global scope.
//func (e *Env) DefineGlobal(k string, v any) error { return e.defineGlobal(k, v) }

// DefineGlobalValue defines symbol in global scope.
func (e *Env) DefineGlobalValue(k string, v reflect.Value) error { return e.defineGlobalValue(k, v) }

// DefineGlobalType defines type in global scope.
//func (e *Env) DefineGlobalType(k string, t any) error { return e.defineGlobalType(k, t) }

// DefineGlobalReflectType defines type in global scope.
//func (e *Env) DefineGlobalReflectType(k string, t reflect.Type) error {
//	return e.defineGlobalReflectType(k, t)
//}

// Define defines symbol in current scope.
func (e *Env) Define(k string, v any) error { return e.define(k, v) }

// DefineCtx is like Define but the running context will be injected at runtime
// as argument to v if v is a function
func (e *Env) DefineCtx(k string, v any) error { return e.defineCtx(k, v) }

// DefineValue defines symbol in current scope.
func (e *Env) DefineValue(k string, v reflect.Value) error { return e.defineValue(k, v) }

// Delete deletes symbol in current scope.
func (e *Env) Delete(k string) error { return e.delete(k) }

// DeleteGlobal deletes the first matching symbol found in current or parent scope.
func (e *Env) DeleteGlobal(k string) error { return e.deleteGlobal(k) }

// DefineType defines type in current scope.
func (e *Env) DefineType(k string, t any) error { return e.defineType(k, t) }

// DefineReflectType defines type in current scope.
func (e *Env) DefineReflectType(k string, t reflect.Type) error { return e.defineReflectType(k, t) }

// Copy the state of the virtual machine environment
//func (e *Env) Copy() *Env { return e.copy() }

// DeepCopy copy recursively the state of the virtual machine environment
func (e *Env) DeepCopy() IEnv { return e.deepCopy() }

//-----------------------------------------------------------------------------

func (e *Env) defineGlobal(k string, v any) error {
	return e.getRootEnv().define(k, v)
}

func (e *Env) defineGlobalValue(k string, v reflect.Value) error {
	return e.getRootEnv().defineValue(k, v)
}

func (e *Env) defineGlobalType(k string, t any) error {
	return e.getRootEnv().DefineType(k, t)
}

func (e *Env) defineGlobalReflectType(k string, t reflect.Type) error {
	return e.getRootEnv().defineReflectType(k, t)
}

func isSymbolNameValid(name string) bool {
	return !strings.Contains(name, ".")
}

type ErrUnknownSymbol struct{ name string }

func (e *ErrUnknownSymbol) Error() string {
	return fmt.Sprintf("unknown symbol '%s'", e.name)
}

func newUnknownSymbolErr(name string) *ErrUnknownSymbol {
	return &ErrUnknownSymbol{name: name}
}

func newEnv() *Env {
	return &Env{
		parent: nil,
		name:   mtx.NewRWMtxPtr(""),
		values: mtx.NewRWMapPtr(map[string]reflect.Value{}),
		types:  mtx.NewRWMapPtr(map[string]reflect.Type{}),
		defers: mtx.NewRWSlicePtr([]CapturedFunc{}),
	}
}

func (e *Env) newEnv() *Env {
	env := newEnv()
	env.parent = e
	e.incrChildCount(1)
	return env
}

func (e *Env) destroy() {
	e.incrChildCount(-1)
}

func (e *Env) withNewEnv(clb func(IEnv)) {
	newenv := e.newEnv()
	defer newenv.destroy()
	clb(newenv)
}

func (e *Env) childCount() int64 {
	return atomic.LoadInt64(&e.childCountVar)
}

func (e *Env) incrChildCount(diff int) {
	atomic.AddInt64(&e.childCountVar, int64(diff))
	if e.parent != nil {
		e.parent.incrChildCount(diff)
	}
}

// try to find a module by name in current env, returns nil if not found
func (e *Env) findModuleInCurrentEnv(name string) *Env {
	if value, ok := e.values.Get(name); ok {
		if foundEnv, ok := value.Interface().(*Env); ok {
			return foundEnv
		}
	}
	return nil
}

// try to find a module by name in the env or any parent, returns nil if not found
func (e *Env) findModule(name string) *Env {
	currEnv := e
	for {
		if module := currEnv.findModuleInCurrentEnv(name); module != nil {
			return module
		}
		// module not found in current env, try parent
		currEnv = currEnv.parent
		if currEnv == nil {
			return nil
		}
	}
}

func (e *Env) getEnvFromPath(path []string) (*Env, error) {
	out := e
	if len(path) < 1 {
		return out, nil
	}
	for i := 0; i < len(path); i++ {
		moduleName := path[i]
		out = utils.Ternary(i == 0, out.findModule, out.findModuleInCurrentEnv)(moduleName)
		if out == nil {
			return nil, fmt.Errorf("no namespace called: %v", moduleName)
		}
	}
	return out, nil
}

func (e *Env) newModule(symbol string) (*Env, error) {
	return e.addPackageHelper(symbol, nil, nil)
}

func (e *Env) addPackage(name string, methods packages.PackageMap, types packages.PackageMap) (*Env, error) {
	return e.addPackageHelper(name, methods, types)
}

func (e *Env) addPackageHelper(name string, methods packages.PackageMap, types packages.PackageMap) (*Env, error) {
	module := e.newEnv()
	destroyFn := module.destroy
	externalName := strings.ReplaceAll(name, "/", "_")
	if err := e.define(externalName, module); err != nil {
		destroyFn()
		return nil, err
	}
	module.name.Store(name)
	for methodName, methodValue := range methods {
		if err := module.define(methodName, methodValue); err != nil {
			destroyFn()
			return nil, err
		}
	}
	for typeName, typeValue := range types {
		if err := module.defineType(typeName, typeValue); err != nil {
			destroyFn()
			return nil, err
		}
	}
	return module, nil
}

func (e *Env) getName() string {
	envName := e.name.Load()
	return utils.TernaryZ(envName, fmt.Sprintf("module<%s>", envName), "n/a")
}

func (e *Env) string() string {
	replaceInterface := func(in string) string { return strings.ReplaceAll(in, "interface {}", "any") }
	valuesArr := make([][]string, 0)
	e.values.Each(func(symbol string, value reflect.Value) {
		if value.Kind() == reflect.Ptr {
			if value.IsValid() && value.CanInterface() {
				if ee, ok := value.Interface().(*Env); ok {
					valuesArr = append(valuesArr, []string{symbol, ee.getName()})
					return
				}
			}
		}
		if value.Kind() != reflect.Func {
			valuesArr = append(valuesArr, []string{symbol, fmt.Sprintf("%#v", value)})
			return
		}
		if value.Kind() == reflect.Func {
			valuesArr = append(valuesArr, []string{symbol, vmUtils.FormatValue(value)})
		}
	})

	typesArr := make([][]string, 0)
	e.types.Each(func(symbol string, aType reflect.Type) {
		aTypeStr := replaceInterface(aType.String())
		aTypeKindStr := aType.Kind().String()
		str := aTypeStr
		if aTypeKindStr != aTypeStr {
			str += " " + aTypeKindStr
		}
		typesArr = append(typesArr, []string{symbol, str})
	})

	sort.Slice(valuesArr, func(i, j int) bool { return valuesArr[i][0] < valuesArr[j][0] })
	sort.Slice(typesArr, func(i, j int) bool { return typesArr[i][0] < typesArr[j][0] })
	maxSymbolLen := 0
	allValsTypes := append(valuesArr, typesArr...)
	for _, v := range allValsTypes {
		maxSymbolLen = max(maxSymbolLen, len(v[0]))
	}

	var buffer bytes.Buffer
	parentStr := utils.Ternary(e.parent == nil, "No parent\n", "Has parent\n")
	buffer.WriteString(parentStr)
	format := "%-" + strconv.Itoa(maxSymbolLen) + "v = %s\n"
	for _, v := range valuesArr {
		buffer.WriteString(fmt.Sprintf(format, v[0], v[1]))
	}
	if len(typesArr) > 0 {
		buffer.WriteString("\n")
		for _, v := range typesArr {
			buffer.WriteString(fmt.Sprintf(format, v[0], v[1]))
		}
	}
	return buffer.String()
}

var nilValue = reflect.New(reflect.TypeOf((*any)(nil)).Elem()).Elem()
var nilType = reflect.TypeOf(nil)

func (e *Env) addr(k string) (reflect.Value, error) {
	if v, ok := e.values.Get(k); ok {
		if v.CanAddr() {
			return v.Addr(), nil
		}
		return nilValue, fmt.Errorf("unaddressable")
	}
	if e.parent == nil {
		return nilValue, fmt.Errorf("undefined symbol '%s'", k)
	}
	return e.parent.addr(k)
}

func (e *Env) typ(k string) (reflect.Type, error) {
	if v, ok := e.types.Get(k); ok {
		return v, nil
	}
	if e.parent == nil {
		if reflectType, ok := basicTypes[k]; ok {
			return reflectType, nil
		}
		return nilType, fmt.Errorf("undefined type '%s'", k)
	}
	return e.parent.typ(k)
}

func (e *Env) get(k string) (any, error) {
	rv, err := e.GetValue(k)
	if !rv.IsValid() || !rv.CanInterface() {
		return nil, err
	}
	return rv.Interface(), err
}

func (e *Env) getValue(k string) (reflect.Value, error) {
	if v, ok := e.values.Get(k); ok {
		return v, nil
	}
	if e.parent == nil {
		return nilValue, fmt.Errorf("undefined symbol '%s'", k)
	}
	return e.parent.getValue(k)
}

// Set modifies value which specified as symbol. It goes to upper scope until
// found or returns error.
func (e *Env) set(k string, v any) error {
	val := nilValue
	if v != nil {
		if rv, ok := v.(reflect.Value); !ok {
			val = reflect.ValueOf(v)
		} else {
			val = rv
		}
	}
	return e.setValue(k, val)
}

func (e *Env) setValue(k string, v reflect.Value) error {
	if e.values.ContainsKey(k) {
		e.values.Insert(k, v)
		return nil
	}
	if e.parent == nil {
		return newUnknownSymbolErr(k)
	}
	return e.parent.setValue(k, v)
}

func (e *Env) define(k string, v any) error {
	val := nilValue
	if v != nil {
		val = reflect.ValueOf(v)
	}
	return e.defineValue(k, val)
}

// InjectCtx is a special type that instruct the VM that it needs to
// inject the running ctx as argument to the function being called.
type InjectCtx struct{ Value any }

func (e *Env) defineCtx(k string, v any) error {
	val := nilValue
	if v != nil {
		if reflect.TypeOf(v).Kind() == reflect.Func {
			val = reflect.ValueOf(&InjectCtx{v})
		} else {
			val = reflect.ValueOf(v)
		}
	}
	return e.defineValue(k, val)
}

func (e *Env) defineValue(k string, v reflect.Value) error {
	if !isSymbolNameValid(k) {
		return newUnknownSymbolErr(k)
	}
	e.values.Insert(k, v)
	return nil
}

func (e *Env) deleteGlobal(k string) error {
	if e.parent == nil || e.values.ContainsKey(k) {
		return e.delete(k)
	}
	return e.parent.deleteGlobal(k)
}

func (e *Env) delete(k string) error {
	if !isSymbolNameValid(k) {
		return newUnknownSymbolErr(k)
	}
	e.values.Delete(k)
	return nil
}

func (e *Env) getRootEnv() (root *Env) {
	root = e
	for root.parent != nil {
		root = root.parent
	}
	return
}

func (e *Env) defineType(k string, t any) error {
	var typ reflect.Type
	if t == nil {
		typ = nilType
	} else {
		var ok bool
		typ, ok = t.(reflect.Type)
		if !ok {
			typ = reflect.TypeOf(t)
		}
	}
	return e.defineReflectType(k, typ)
}

func (e *Env) defineReflectType(k string, t reflect.Type) error {
	if !isSymbolNameValid(k) {
		return newUnknownSymbolErr(k)
	}
	e.types.Insert(k, t)
	return nil
}

func (e *Env) copy() *Env {
	copyEnv := newEnv()
	copyEnv.parent = e.parent
	copyEnv.values.Store(e.values.Clone())
	if e.types != nil {
		copyEnv.types.Store(e.types.Clone())
	}
	return copyEnv
}

func (e *Env) deepCopy() *Env {
	copyEnv := e.copy()
	if copyEnv.parent != nil {
		copyEnv.parent = copyEnv.parent.deepCopy()
	}
	return copyEnv
}
