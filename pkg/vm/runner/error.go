package runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/alaingilbert/anko/pkg/ast"
	vmUtils "github.com/alaingilbert/anko/pkg/vm/utils"
	"reflect"
)

// Error provides a convenient interface for handling runtime error.
// It can be Error interface with type cast which can call Pos().
type Error struct {
	Message string
	Pos     ast.Position
	cause   error
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.cause
}

type IsVmFunc struct{ context.Context }

var (
	StringType         = reflect.TypeOf("a")
	interfaceType      = reflect.ValueOf([]any{int64(1)}).Index(0).Type()
	InterfaceSliceType = reflect.TypeOf([]any{})
	reflectValueType   = reflect.TypeOf(reflect.Value{})
	errorType          = reflect.TypeOf((*error)(nil)).Elem()
	vmErrorType        = reflect.TypeOf(&Error{})
	isVmFuncType       = reflect.TypeOf((*IsVmFunc)(nil))

	nilValue                  = vmUtils.NilValue
	trueValue                 = reflect.ValueOf(true)
	falseValue                = reflect.ValueOf(false)
	zeroValue                 = reflect.Value{}
	reflectValueErrorNilValue = reflect.ValueOf(reflect.New(errorType).Elem())

	ErrUnknownStmt              = errors.New("unknown statement")
	ErrUnknownExpr              = errors.New("unknown expression")
	ErrUnknownOperator          = errors.New("unknown operator")
	ErrInvalidSliceIndex        = errors.New("invalid slice index")
	ErrInvalidTypeConversion    = errors.New("invalid type conversion")
	ErrIndexOutOfRange          = errors.New("index out of range")
	ErrIndexMustBeNumber        = errors.New("index must be a number")
	ErrNoSupportMemberOpInvalid = NewNoSupportMemberOpError("invalid")

	// ErrBreak when there is an unexpected break statement
	ErrBreak = errors.New("unexpected break statement")
	// ErrContinue when there is an unexpected continue statement
	ErrContinue = errors.New("unexpected continue statement")
	// ErrReturn when there is an unexpected return statement
	ErrReturn = errors.New("unexpected return statement")
)

func getPos(pos ast.Pos) ast.Position {
	out := ast.Position{Line: 1, Column: 1}
	if pos != nil {
		out = pos.Position()
	}
	return out
}

// newError makes error interface with message.
// This doesn't overwrite last error.
func newError(pos ast.Pos, err error) error {
	var vmErr *Error
	if errors.As(err, &vmErr) {
		return err
	}
	return &Error{Message: err.Error(), Pos: getPos(pos), cause: err}
}

// newStringError makes error interface with message.
func newStringError(pos ast.Pos, err string) error {
	return newError(pos, errors.New(err))
}

// newErrorf makes error interface with message.
func newErrorf(pos ast.Pos, format string, args ...any) error {
	return newStringError(pos, fmt.Sprintf(format, args...))
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}

type WrongArgTypeError struct {
	want     string
	received string
}

func NewWrongArgTypeError(want, receive string) *WrongArgTypeError {
	return &WrongArgTypeError{want: want, received: receive}
}

func (e *WrongArgTypeError) Error() string {
	return "function wants argument type " + e.want + " but received type " + e.received
}

type NoSupportMemberOpError struct {
	typ string
}

func NewNoSupportMemberOpError(typ string) *NoSupportMemberOpError {
	return &NoSupportMemberOpError{typ: typ}
}

func (e *NoSupportMemberOpError) Error() string {
	return "type " + e.typ + " does not support member operation"
}

type CannotCallError struct {
	typ string
}

func NewCannotCallError(typ string) *CannotCallError {
	return &CannotCallError{typ: typ}
}

func (e *CannotCallError) Error() string {
	return "cannot call type " + e.typ
}

type TypeCannotBeAssignedError struct {
	from, to, into string
}

func NewTypeCannotBeAssignedError(from, to, into string) *TypeCannotBeAssignedError {
	return &TypeCannotBeAssignedError{from, to, into}
}

func (e *TypeCannotBeAssignedError) Error() string {
	out := fmt.Sprintf("type %s cannot be assigned to type %s", e.from, e.to)
	if e.into != "" {
		out += fmt.Sprintf(" for %s", e.into)
	}
	return out
}
