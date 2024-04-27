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

func ConvertToInt(oType ObjectType, val int64) Object {
	var obj Object
	switch oType {
	case INT_OBJ:
		obj = &Int{Value: int(val)}
	case INT8_OBJ:
		obj = &Int8{Value: int8(val)}
	case INT16_OBJ:
		obj = &Int16{Value: int16(val)}
	case INT32_OBJ:
		obj = &Int32{Value: int32(val)}
	case INT64_OBJ:
		obj = &Int64{Value: val}
	case UINT_OBJ:
		obj = &Uint{Value: uint(val)}
	case UINT8_OBJ:
		obj = &Uint8{Value: uint8(val)}
	case UINT16_OBJ:
		obj = &Uint16{Value: uint16(val)}
	case UINT32_OBJ:
		obj = &Uint32{Value: uint32(val)}
	case UINT64_OBJ:
		obj = &Uint64{Value: uint64(val)}
	default:
		obj = NewError("integer no support %s", oType)
	}
	return obj
}

func ConvertToFloat(oType ObjectType, val float64) Object {
	var obj Object
	switch oType {
	case FLOAT32_OBJ:
		obj = &Float32{Value: float32(val)}
	case FLOAT64_OBJ:
		obj = &Float64{Value: val}
	default:
		obj = NewError("float no support %s", oType)
	}
	return obj
}

func ConvertToBoolean(b bool) Object {
	if b {
		return TRUE
	} else {
		return FALSE
	}
}
