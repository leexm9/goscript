package object

import "fmt"

var (
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NULL  = &Null{}
)

func NewError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func IsError(obj Object) bool {
	return obj.Type() == ERROR_OBJ
}
