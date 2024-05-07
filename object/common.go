package object

import (
	"fmt"
	"go/ast"
)

var (
	TRUE     = &Boolean{Value: true}
	FALSE    = &Boolean{Value: false}
	NULL     = &Null{}
	CONTINUE = &Continue{}
	BREAK    = &Break{}
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

func GetDefaultValueWithExpr(expr ast.Expr) Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return GetDefaultObject(expr.Name)
	case *ast.ArrayType:
		elem := GetDefaultValueWithExpr(expr.Elt)
		return &Array{ElemType: elem.Type()}
	case *ast.MapType:
		key := GetDefaultValueWithExpr(expr.Key)
		value := GetDefaultValueWithExpr(expr.Value)
		return &Hash{KeyType: key.Type(), ValueType: value.Type()}
	default:
		return NewError("GetDefaultValueWithExpr not support %T", expr)
	}
}

func GetDefaultObject(objType string) Object {
	switch objType {
	case "int":
		return &Int{Value: 0}
	case "int8":
		return &Int8{Value: 0}
	case "int16":
		return &Int16{Value: 0}
	case "int32":
		return &Int32{Value: 0}
	case "int64":
		return &Int64{Value: 0}
	case "uint":
		return &Uint{Value: 0}
	case "uint8":
		return &Uint8{Value: 0}
	case "uint16":
		return &Uint16{Value: 0}
	case "uint32":
		return &Uint32{Value: 0}
	case "uint64":
		return &Uint64{Value: 0}
	case "float32":
		return &Float32{Value: 0.0}
	case "float64":
		return &Float64{Value: 0.0}
	case "byte":
		return &Byte{Value: 0}
	case "rune":
		return &Rune{Value: 0}
	case "string":
		return &String{Value: ""}
	case "bool":
		return &Boolean{Value: false}
	default:
		return NewError("not known type: %s", objType)
	}
}

func ConvertValueWithType(obj, typeObj Object) Object {
	if typeObj == nil {
		return obj
	}
	toType := typeObj.Type()
	if toType.IsInteger() {
		if tmp, ok := obj.(Integer); ok {
			return ConvertToInt(toType, tmp.Integer())
		} else {
			return NewError("cannot convert (untyped '%s' constant) to type %s", obj.Type(), toType)
		}
	} else if toType.IsFloat() {
		if tmp, ok := obj.(Float); ok {
			return ConvertToFloat(toType, tmp.Float())
		} else {
			return NewError("cannot convert (untyped '%s' constant) to type %s", obj.Type(), toType)
		}
	} else if toType == STRING_OBJ {
		if obj.Type() == STRING_OBJ {
			return obj
		} else {
			return NewError("cannot convert (untyped '%s' constant) to type string", obj.Type())
		}
	} else if toType == BOOLEAN_OBJ {
		if obj.Type() == BOOLEAN_OBJ {
			return obj
		} else {
			return NewError("cannot convert (untyped '%s' constant) to type bool", obj.Type())
		}
	} else if toType == ARRAY_OBJ {
		array := &Array{ElemType: typeObj.(*Array).ElemType}
		array.Elements = []Object{}
		if obj != nil {
			defObj := GetDefaultObject(array.ElemType.String())
			for _, elem := range obj.(*Array).Elements {
				tmp := ConvertValueWithType(elem, defObj)
				array.Elements = append(array.Elements, tmp)
			}
		}
		return array
	} else if toType == HASH_OBJ {
		hash := &Hash{KeyType: typeObj.(*Hash).KeyType, ValueType: typeObj.(*Hash).ValueType}
		hash.Pairs = map[HashKey]HashPair{}
		if obj != nil {
			defKObj := GetDefaultObject(hash.KeyType.String())
			defVObj := GetDefaultObject(hash.ValueType.String())
			for _, pair := range obj.(*Hash).Pairs {
				k := ConvertValueWithType(pair.Key, defKObj)
				v := ConvertValueWithType(pair.Value, defVObj)
				hash.Pairs[k.(Hashable).HashKey()] = HashPair{Key: k, Value: v}
			}
		}
		return hash
	}
	return obj
}

func GetDefaultValueFromElem(elemType ElemType) Object {
	switch elemType.TypeElem {
	case ElemBase:
		return GetDefaultObject(elemType.Type.Name)
	case ElemArray:
		elem := GetDefaultObject(elemType.Type.Name)
		return &Array{ElemType: elem.Type()}
	case ElemHash:
		key := GetDefaultObject(elemType.Types[0].Name)
		value := GetDefaultObject(elemType.Types[1].Name)
		return &Hash{KeyType: key.Type(), ValueType: value.Type()}
	default:
		return nil
	}
}
