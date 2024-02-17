package vm

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	"github.com/alaingilbert/anko/pkg/parser"
	"reflect"
)

type (
	// Error provides a convenient interface for handling runtime error.
	// It can be Error interface with type cast which can call Pos().
	Error struct {
		Message string
		Pos     ast.Position
	}
)

type IsVmFunc struct{ context.Context }

var (
	nilType            = reflect.TypeOf(nil)
	stringType         = reflect.TypeOf("a")
	interfaceType      = reflect.ValueOf([]any{int64(1)}).Index(0).Type()
	interfaceSliceType = reflect.TypeOf([]any{})
	reflectValueType   = reflect.TypeOf(reflect.Value{})
	errorType          = reflect.TypeOf((*error)(nil)).Elem()
	vmErrorType        = reflect.TypeOf(&Error{})
	isVmFuncType       = reflect.TypeOf((*IsVmFunc)(nil))

	nilValue                  = reflect.New(reflect.TypeOf((*any)(nil)).Elem()).Elem()
	trueValue                 = reflect.ValueOf(true)
	falseValue                = reflect.ValueOf(false)
	zeroValue                 = reflect.Value{}
	reflectValueErrorNilValue = reflect.ValueOf(reflect.New(errorType).Elem())

	// ErrBreak when there is an unexpected break statement
	ErrBreak = errors.New("unexpected break statement")
	// ErrContinue when there is an unexpected continue statement
	ErrContinue = errors.New("unexpected continue statement")
	// ErrReturn when there is an unexpected return statement
	ErrReturn = errors.New("unexpected return statement")
	// ErrInterrupt when execution has been interrupted
	ErrInterrupt = errors.New("execution interrupted")

	ErrInvalidInput = errors.New("invalid input")
)

// newErrorf makes error interface with message.
func newErrorf(pos ast.Pos, format string, args ...any) error {
	return newStringError(pos, fmt.Sprintf(format, args...))
}

// newError makes error interface with message.
// This doesn't overwrite last error.
func newError(pos ast.Pos, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrBreak) || errors.Is(err, ErrContinue) || errors.Is(err, ErrReturn) {
		return err
	}
	var pe *parser.Error
	if errors.As(err, &pe) {
		return pe
	}
	var ee *Error
	if errors.As(err, &ee) {
		return ee
	}
	return newStringError(pos, err.Error())
}

// newStringError makes error interface with message.
func newStringError(pos ast.Pos, err string) error {
	pos1 := ast.Position{Line: 1, Column: 1}
	if pos != nil {
		pos1 = pos.Position()
	}
	return &Error{Message: err, Pos: pos1}
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}
