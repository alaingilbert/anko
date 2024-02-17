package runner

import (
	"fmt"
	"reflect"
)

// reflectValueSliceToInterfaceSlice convert from a slice of reflect.Value to a interface slice
// returned in normal reflect.Value form
func reflectValueSliceToInterfaceSlice(valueSlice []reflect.Value) reflect.Value {
	interfaceSlice := make([]any, 0, len(valueSlice))
	for _, value := range valueSlice {
		if !value.IsValid() {
			interfaceSlice = append(interfaceSlice, nil)
			continue
		}
		if value.Kind() == reflect.Interface && !value.IsNil() {
			value = value.Elem()
		}
		if value.CanInterface() {
			interfaceSlice = append(interfaceSlice, value.Interface())
		} else {
			interfaceSlice = append(interfaceSlice, nil)
		}
	}
	return reflect.ValueOf(interfaceSlice)
}

// convertReflectValueToType trys to covert the reflect.Value to the reflect.Type
// if it can not, it returns the original rv and an error
func convertReflectValueToType(vmp *VmParams, rv reflect.Value, rt reflect.Type) (reflect.Value, error) {
	if !rv.IsValid() {
		// if not valid return a valid reflect.Value of the reflect.Type
		return MakeValue(rt)
	}
	if rt == interfaceType || rv.Type() == rt {
		// if reflect.Type is interface or the types match, return the provided reflect.Value
		return rv, nil
	}
	if rv.Type().ConvertibleTo(rt) {
		// if reflect can covert, do that conversion and return
		return rv.Convert(rt), nil
	}
	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) &&
		(rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array) {
		// covert slice or array
		return convertSliceOrArray(vmp, rv, rt)
	}
	if rv.Kind() == rt.Kind() {
		// kind matches
		switch rv.Kind() {
		case reflect.Map:
			// convert map
			return convertMap(vmp, rv, rt)
		case reflect.Func:
			// for runVMFunction conversions, call convertVMFunctionToType
			return convertVMFunctionToType(vmp, rv, rt)
		case reflect.Pointer:
			// both rv and rt are pointers, convert what they are pointing to
			value, err := convertReflectValueToType(vmp, rv.Elem(), rt.Elem())
			if err != nil {
				return rv, err
			}
			// need to make a new value to be able to set it
			ptrV, err := MakeValue(rt)
			if err != nil {
				return rv, err
			}
			// set value and return new pointer
			ptrV.Elem().Set(value)
			return ptrV, nil
		}
	}
	if rv.Type() == interfaceType {
		if rv.IsNil() {
			// return nil of correct type
			return reflect.Zero(rt), nil
		}
		// try to convert the element
		return convertReflectValueToType(vmp, rv.Elem(), rt)
	}

	// TODO: need to handle the case where either rv or rt are a pointer but not both

	return rv, fmt.Errorf("invalid type conversion")
}

// convertSliceOrArray trys to covert the reflect.Value slice or array to the slice or array reflect.Type
func convertSliceOrArray(vmp *VmParams, rv reflect.Value, rt reflect.Type) (reflect.Value, error) {
	rtElemType := rt.Elem()

	// try to covert elements to new slice/array
	var value reflect.Value
	if rt.Kind() == reflect.Slice {
		// make slice
		value = reflect.MakeSlice(rt, rv.Len(), rv.Len())
	} else {
		// make array
		value = reflect.New(rt).Elem()
	}

	var err error
	var v reflect.Value
	for i := 0; i < rv.Len(); i++ {
		if !value.Index(i).CanSet() {
			// is there a way for new slice/array not to be settable?
			return rv, fmt.Errorf("invalid type conversion")
		}
		v, err = convertReflectValueToType(vmp, rv.Index(i), rtElemType)
		if err != nil {
			return rv, err
		}
		value.Index(i).Set(v)
	}

	// return new converted slice or array
	return value, nil
}

// convertMap trys to covert the reflect.Value map to the map reflect.Type
func convertMap(vmp *VmParams, rv reflect.Value, rt reflect.Type) (reflect.Value, error) {
	rtKey := rt.Key()
	rtElem := rt.Elem()

	// create new map
	// note creating slice as work around to create map
	// just doing MakeMap can give incorrect type for defined types
	newMap := reflect.MakeSlice(reflect.SliceOf(rt), 0, 1)
	newMap = reflect.Append(newMap, reflect.MakeMap(reflect.MapOf(rtKey, rtElem))).Index(0)

	// copy keys to new map
	// The only way to do this right now is to get all the keys.
	// At some point there will be a MapRange that could be used.
	// https://github.com/golang/go/issues/11104
	// In the meantime using MapKeys, which will costly for large maps.

	it := rv.MapRange()
	for mapIterNext(it, vmp) {
		oldKey := it.Key()
		oldVal := it.Value()
		newKey, err := convertReflectValueToType(vmp, oldKey, rtKey)
		if err != nil {
			return rv, err
		}
		oldVal, err = convertReflectValueToType(vmp, oldVal, rtElem)
		if err != nil {
			return rv, err
		}
		newMap.SetMapIndex(newKey, oldVal)
	}
	return newMap, nil
}

// convertVMFunctionToType is for translating a runVMFunction into the correct type
// so, it can be passed to a Go function argument with the correct static types
// it creates a translation function runVMConvertFunction
func convertVMFunctionToType(vmp *VmParams, rv reflect.Value, rt reflect.Type) (reflect.Value, error) {
	// only translates runVMFunction type
	if !checkIfRunVMFunction(rv.Type()) {
		return rv, fmt.Errorf("invalid type conversion")
	}

	// create runVMConvertFunction to match reflect.Type
	// this function is being called by the Go function
	runVMConvertFunction := func(in []reflect.Value) []reflect.Value {
		// note: this function is being called by another reflect Call
		// only way to pass along any errors is by panic

		rtNumIn := rt.NumIn()
		rtNumOut := rt.NumOut()
		// make the reflect.Value slice of each of the VM reflect.Value
		args := make([]reflect.Value, 0, rtNumIn+1)
		// for runVMFunction first arg is always IsVmFunc & context
		// TOFIX: use normal context
		args = append(args, reflect.ValueOf(&IsVmFunc{vmp.ctx}))
		for i := 0; i < rtNumIn; i++ {
			// have to do the double reflect.ValueOf that runVMFunction expects
			args = append(args, reflect.ValueOf(in[i]))
		}

		// Call runVMFunction
		rvs := rv.Call(args)

		// call processCallReturnValues to process runVMFunction return values
		// returns normal VM reflect.Value form
		rv, err := processCallReturnValues(rvs, true, false)
		if err != nil {
			panic(err)
		}

		if rtNumOut < 1 {
			// Go function does not want any return values, so give it none
			return []reflect.Value{}
		}
		if rtNumOut < 2 {
			// Go function wants one return value
			// will try to covert to reflect.Value correct type and return
			rv, err = convertReflectValueToType(vmp, rv, rt.Out(0))
			if err != nil {
				panic("function wants return type " + rt.Out(0).String() + " but received type " + rv.Type().String())
			}
			return []reflect.Value{rv}
		}

		// Go function wants more than one return value
		// make sure we have a slice/array with enough values

		if !rv.IsValid() {
			panic(fmt.Sprintf("function wants %v return values but received invalid", rtNumOut))
		}
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			panic(fmt.Sprintf("function wants %v return values but received %v", rtNumOut, rv.Kind().String()))
		}
		if rv.Len() < rtNumOut {
			panic(fmt.Sprintf("function wants %v return values but received %v values", rtNumOut, rv.Len()))
		}

		// try to covert each value in slice to wanted type and put into a reflect.Value slice
		rvs = make([]reflect.Value, rtNumOut)
		for i := 0; i < rv.Len(); i++ {
			rvs[i], err = convertReflectValueToType(vmp, rv.Index(i), rt.Out(i))
			if err != nil {
				panic("function wants return type " + rt.Out(i).String() + " but received type " + rvs[i].Type().String())
			}
		}

		// return created reflect.Value slice
		return rvs
	}

	// make the reflect.Value function that calls runVMConvertFunction
	return reflect.MakeFunc(rt, runVMConvertFunction), nil
}
