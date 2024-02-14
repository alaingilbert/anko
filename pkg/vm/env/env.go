package env

import (
	"bytes"
	"fmt"
	"github.com/alaingilbert/mtx"
	"reflect"
	"strings"
)

// Env provides interface to run VM. This mean function scope and blocked-scope.
// If stack goes to blocked-scope, it will make new Env.
type Env struct {
	parent *Env
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
	{name: "bool", typ: reflect.TypeOf(true)},
	{name: "string", typ: reflect.TypeOf("a")},
	{name: "int", typ: reflect.TypeOf(int(1))},
	{name: "int32", typ: reflect.TypeOf(int32(1))},
	{name: "int64", typ: reflect.TypeOf(int64(1))},
	{name: "uint", typ: reflect.TypeOf(uint(1))},
	{name: "uint32", typ: reflect.TypeOf(uint32(1))},
	{name: "uint64", typ: reflect.TypeOf(uint64(1))},
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
		values: mtx.NewMapPtr[string, reflect.Value](nil),
		types:  mtx.NewMapPtr[string, reflect.Type](nil),
		defers: mtx.NewSlicePtr[CapturedFunc](nil),
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
	return module, e.Define(symbol, module)
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

// String returns string of values and types in current scope.
func (e *Env) String() string {
	var buffer bytes.Buffer
	if e.parent == nil {
		buffer.WriteString("No parent\n")
	} else {
		buffer.WriteString("Has parent\n")
	}
	e.values.Each(func(symbol string, value reflect.Value) {
		buffer.WriteString(fmt.Sprintf("%v = %#v\n", symbol, value))
	})
	e.types.Each(func(symbol string, aType reflect.Type) {
		buffer.WriteString(fmt.Sprintf("%v = %v\n", symbol, aType))
	})
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
