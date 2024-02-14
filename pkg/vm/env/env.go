package env

import (
	"bytes"
	"fmt"
	"github.com/alaingilbert/anko/pkg/utils"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"github.com/alaingilbert/mtx"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Env provides interface to run VM. This mean function scope and blocked-scope.
// If stack goes to blocked-scope, it will make new Env.
type Env struct {
	parent *Env
	name   *mtx.RWMtx[string]
	values *mtx.Map[string, reflect.Value]
	types  *mtx.Map[string, reflect.Type]
	defers *mtx.Slice[CapturedFunc]
}

func (e *Env) Values() *mtx.Map[string, reflect.Value] {
	return e.values
}

func (e *Env) Types() *mtx.Map[string, reflect.Type] {
	return e.types
}

func (e *Env) Defers() *mtx.Slice[CapturedFunc] {
	return e.defers
}

var basicTypes = []struct {
	name string
	typ  reflect.Type
}{
	{name: "interface", typ: reflect.ValueOf([]any{int64(1)}).Index(0).Type()},
	{name: "any", typ: reflect.ValueOf([]any{int64(1)}).Index(0).Type()},
	{name: "bool", typ: reflect.TypeOf(true)},
	{name: "string", typ: reflect.TypeOf("a")},
	{name: "int", typ: reflect.TypeOf(int(1))},
	{name: "int8", typ: reflect.TypeOf(int8(1))},
	{name: "int16", typ: reflect.TypeOf(int16(1))},
	{name: "int32", typ: reflect.TypeOf(int32(1))},
	{name: "int64", typ: reflect.TypeOf(int64(1))},
	{name: "uint", typ: reflect.TypeOf(uint(1))},
	{name: "uint8", typ: reflect.TypeOf(uint8(1))},
	{name: "uint16", typ: reflect.TypeOf(uint16(1))},
	{name: "uint32", typ: reflect.TypeOf(uint32(1))},
	{name: "uint64", typ: reflect.TypeOf(uint64(1))},
	{name: "uintptr", typ: reflect.TypeOf(uintptr(1))},
	{name: "byte", typ: reflect.TypeOf(byte(1))},
	{name: "rune", typ: reflect.TypeOf('a')},
	{name: "float32", typ: reflect.TypeOf(float32(1))},
	{name: "float64", typ: reflect.TypeOf(float64(1))},
}

func newBasicTypes() map[string]reflect.Type {
	types := make(map[string]reflect.Type, len(basicTypes))
	for i := 0; i < len(basicTypes); i++ {
		types[basicTypes[i].name] = basicTypes[i].typ
	}
	return types
}

// CapturedFunc is stacked in the scope
type CapturedFunc struct {
	Func      reflect.Value
	Args      []reflect.Value
	CallSlice bool
}

func newEnv() *Env {
	return &Env{
		parent: nil,
		name:   mtx.NewRWMtxPtr(""),
		values: mtx.NewRWMapPtr[string, reflect.Value](nil),
		types:  mtx.NewRWMapPtr[string, reflect.Type](nil),
		defers: mtx.NewRWSlicePtr[CapturedFunc](nil),
	}
}

// NewEnv creates new global scope.
func NewEnv() *Env {
	env := newEnv()
	env.types.Store(newBasicTypes())
	return env
}

// NewEnv creates new child scope.
func (e *Env) NewEnv() *Env {
	env := newEnv()
	env.parent = e
	return env
}

// NewModule creates new child scope and define it as a symbol.
// This is a shortcut for calling e.NewEnv then Define that new Env.
func (e *Env) NewModule(symbol string) (*Env, error) {
	module := e.NewEnv()
	if err := e.Define(symbol, module); err != nil {
		return nil, err
	}
	module.name.Store(symbol)
	return module, nil
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

// GetEnvFromPath returns Env from path
func (e *Env) GetEnvFromPath(path []string) (*Env, error) {
	out := e
	if len(path) < 1 {
		return out, nil
	}
	buildErr := func(name string) error {
		return fmt.Errorf("no namespace called: %v", name)
	}
	for i := 0; i < len(path); i++ {
		moduleName := path[i]
		out = utils.Ternary(i == 0, out.findModule, out.findModuleInCurrentEnv)(moduleName)
		if out == nil {
			return nil, buildErr(moduleName)
		}
	}
	return out, nil
}

func isSymbolNameValid(name string) bool {
	return !strings.Contains(name, ".")
}

type ErrUnknownSymbol struct {
	name string
}

func (e *ErrUnknownSymbol) Error() string {
	return fmt.Sprintf("unknown symbol '%s'", e.name)
}

func newUnknownSymbol(name string) *ErrUnknownSymbol {
	return &ErrUnknownSymbol{name: name}
}

// AddPackage creates a new env with a name that has methods and types in it. Created under the parent env
func (e *Env) AddPackage(name string, methods map[string]any, types map[string]any) (*Env, error) {
	if !isSymbolNameValid(name) {
		return nil, newUnknownSymbol(name)
	}
	var err error
	pack := e.NewEnv()

	for methodName, methodValue := range methods {
		err = pack.Define(methodName, methodValue)
		if err != nil {
			return pack, err
		}
	}
	for typeName, typeValue := range types {
		err = pack.DefineType(typeName, typeValue)
		if err != nil {
			return pack, err
		}
	}

	// can ignore error from Define because of check at the start of the function
	_ = e.Define(name, pack)
	return pack, nil
}

func (e *Env) Name() string {
	envName := e.name.Load()
	return utils.TernaryZ(envName, fmt.Sprintf("module<%s>", envName), "n/a")
}

// String returns string of values and types in current scope.
func (e *Env) String() string {
	replaceInterface := func(in string) string { return strings.ReplaceAll(in, "interface {}", "any") }
	valuesArr := make([][]string, 0)
	e.values.Each(func(symbol string, value reflect.Value) {
		if value.Kind() == reflect.Ptr {
			if value.IsValid() && value.CanInterface() {
				if ee, ok := value.Interface().(*Env); ok {
					valuesArr = append(valuesArr, []string{symbol, ee.Name()})
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
		typesArr = append(typesArr, []string{symbol, fmt.Sprintf("%s", replaceInterface(aType.String()))})
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
	for _, v := range allValsTypes {
		buffer.WriteString(fmt.Sprintf(format, v[0], v[1]))
	}
	return buffer.String()
}

var nilValue = reflect.New(reflect.TypeOf((*any)(nil)).Elem()).Elem()
var nilType = reflect.TypeOf(nil)

// Addr returns pointer value which specified symbol. It goes to upper scope until
// found or returns error.
func (e *Env) Addr(k string) (reflect.Value, error) {
	if v, ok := e.values.Get(k); ok {
		if v.CanAddr() {
			return v.Addr(), nil
		}
		return nilValue, fmt.Errorf("unaddressable")
	}
	if e.parent == nil {
		return nilValue, fmt.Errorf("undefined symbol '%s'", k)
	}
	return e.parent.Addr(k)
}

// Type returns type which specified symbol. It goes to upper scope until
// found or returns error.
func (e *Env) Type(k string) (reflect.Type, error) {
	if v, ok := e.types.Get(k); ok {
		return v, nil
	}
	if e.parent == nil {
		return nilType, fmt.Errorf("undefined type '%s'", k)
	}
	return e.parent.Type(k)
}

// Get returns value which specified symbol. It goes to upper scope until
// found or returns error.
func (e *Env) Get(k string) (any, error) {
	rv, err := e.GetValue(k)
	if !rv.IsValid() || !rv.CanInterface() {
		return nil, err
	}
	return rv.Interface(), err
}

func (e *Env) GetValue(k string) (reflect.Value, error) {
	if v, ok := e.values.Get(k); ok {
		return v, nil
	}
	if e.parent == nil {
		return nilValue, fmt.Errorf("undefined symbol '%s'", k)
	}
	return e.parent.GetValue(k)
}

// Set modifies value which specified as symbol. It goes to upper scope until
// found or returns error.
func (e *Env) Set(k string, v any) error {
	val := nilValue
	if v != nil {
		val = reflect.ValueOf(v)
	}
	return e.SetValue(k, val)
}

func (e *Env) SetValue(k string, v reflect.Value) error {
	_, ok := e.values.Get(k)
	if ok {
		e.values.Insert(k, v)
		return nil
	}
	if e.parent == nil {
		return newUnknownSymbol(k)
	}
	return e.parent.SetValue(k, v)
}

// DefineGlobal defines symbol in global scope.
func (e *Env) DefineGlobal(k string, v any) error {
	return e.getRootEnv().Define(k, v)
}

// DefineGlobalValue defines symbol in global scope.
func (e *Env) DefineGlobalValue(k string, v reflect.Value) error {
	return e.getRootEnv().DefineValue(k, v)
}

// Define defines symbol in current scope.
func (e *Env) Define(k string, v any) error {
	val := nilValue
	if v != nil {
		val = reflect.ValueOf(v)
	}
	return e.DefineValue(k, val)
}

// DefineValue defines symbol in current scope.
func (e *Env) DefineValue(k string, v reflect.Value) error {
	if !isSymbolNameValid(k) {
		return newUnknownSymbol(k)
	}
	e.values.Insert(k, v)
	return nil
}

// Delete deletes symbol in current scope.
func (e *Env) Delete(k string) error {
	if !isSymbolNameValid(k) {
		return newUnknownSymbol(k)
	}
	e.values.Delete(k)
	return nil
}

// DeleteGlobal deletes the first matching symbol found in current or parent scope.
func (e *Env) DeleteGlobal(k string) error {
	if e.parent == nil {
		return e.Delete(k)
	}
	if _, ok := e.values.Get(k); ok {
		return e.Delete(k)
	}
	return e.parent.DeleteGlobal(k)
}

func (e *Env) getRootEnv() (root *Env) {
	root = e
	for root.parent != nil {
		root = root.parent
	}
	return
}

// DefineGlobalType defines type in global scope.
func (e *Env) DefineGlobalType(k string, t any) error {
	return e.getRootEnv().DefineType(k, t)
}

// DefineGlobalReflectType defines type in global scope.
func (e *Env) DefineGlobalReflectType(k string, t reflect.Type) error {
	return e.getRootEnv().DefineReflectType(k, t)
}

// DefineType defines type in current scope.
func (e *Env) DefineType(k string, t any) error {
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

	return e.DefineReflectType(k, typ)
}

// DefineReflectType defines type in current scope.
func (e *Env) DefineReflectType(k string, t reflect.Type) error {
	if !isSymbolNameValid(k) {
		return newUnknownSymbol(k)
	}
	e.types.Insert(k, t)
	return nil
}

// Copy the state of the virtual machine environment
func (e *Env) Copy() *Env {
	copyEnv := newEnv()
	copyEnv.parent = e.parent
	copyEnv.values.Store(e.values.Clone())
	if e.types != nil {
		copyEnv.types.Store(e.types.Clone())
	}
	return copyEnv
}

// DeepCopy copy recursively the state of the virtual machine environment
func (e *Env) DeepCopy() *Env {
	copyEnv := e.Copy()
	if copyEnv.parent != nil {
		copyEnv.parent = copyEnv.parent.DeepCopy()
	}
	return copyEnv
}
