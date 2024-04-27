package object

import "fmt"

var Builtins = []struct {
	Name    string
	rtName  int
	Builtin *Builtin
}{
	{
		"int", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Int{Value: int(arg.(Integer).Integer())}
				case *Int:
					return arg
				case *Uint64:
					return &Int{Value: int(arg.Value)}
				case *Byte:
					return &Int{Value: int(arg.Value)}
				case *Rune:
					return &Int{Value: int(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'int'", args[0].Type())
				}
			},
		},
	},
	{
		"int8", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Int8{Value: int8(arg.(Integer).Integer())}
				case *Int8:
					return arg
				case *Uint64:
					return &Int8{Value: int8(arg.Value)}
				case *Byte:
					return &Int8{Value: int8(arg.Value)}
				case *Rune:
					return &Int8{Value: int8(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'int8'", args[0].Type())
				}
			},
		},
	},
	{
		"int16", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int32, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Int16{Value: int16(arg.(Integer).Integer())}
				case *Int16:
					return arg
				case *Uint64:
					return &Int16{Value: int16(arg.Value)}
				case *Byte:
					return &Int16{Value: int16(arg.Value)}
				case *Rune:
					return &Int16{Value: int16(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'int16'", args[0].Type())
				}
			},
		},
	},
	{
		"int32", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Int32{Value: int32(arg.(Integer).Integer())}
				case *Int32:
					return arg
				case *Uint64:
					return &Int32{Value: int32(arg.Value)}
				case *Byte:
					return &Int32{Value: int32(arg.Value)}
				case *Rune:
					return &Int32{Value: arg.Value}
				default:
					return NewError("cannot convert the type '%s' to type 'int32'", args[0].Type())
				}
			},
		},
	},
	{
		"int64", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Uint, *Uint8, *Uint16, *Uint32:
					return &Int64{Value: arg.(Integer).Integer()}
				case *Int64:
					return arg
				case *Uint64:
					return &Int64{Value: int64(arg.Value)}
				case *Byte:
					return &Int64{Value: int64(arg.Value)}
				case *Rune:
					return &Int64{Value: int64(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'int64'", args[0].Type())
				}
			},
		},
	},
	{
		"uint", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint8, *Uint16, *Uint32:
					return &Uint{Value: uint(arg.(Integer).Integer())}
				case *Uint:
					return arg
				case *Uint64:
					return &Uint{Value: uint(arg.Value)}
				case *Byte:
					return &Uint{Value: uint(arg.Value)}
				case *Rune:
					return &Uint{Value: uint(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'uint'", args[0].Type())
				}
			},
		},
	},
	{
		"uint8", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint16, *Uint32:
					return &Uint8{Value: uint8(arg.(Integer).Integer())}
				case *Uint8:
					return arg
				case *Uint64:
					return &Uint8{Value: uint8(arg.Value)}
				case *Byte:
					return &Uint8{Value: arg.Value}
				case *Rune:
					return &Uint8{Value: uint8(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'uint8'", args[0].Type())
				}
			},
		},
	},
	{
		"uint16", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint32:
					return &Uint16{Value: uint16(arg.(Integer).Integer())}
				case *Uint16:
					return arg
				case *Uint64:
					return &Uint16{Value: uint16(arg.Value)}
				case *Byte:
					return &Uint16{Value: uint16(arg.Value)}
				case *Rune:
					return &Uint16{Value: uint16(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'uint16'", args[0].Type())
				}
			},
		},
	},
	{
		"uint32", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16:
					return &Uint32{Value: uint32(arg.(Integer).Integer())}
				case *Uint32:
					return arg
				case *Uint64:
					return &Uint32{Value: uint32(arg.Value)}
				case *Byte:
					return &Uint32{Value: uint32(arg.Value)}
				case *Rune:
					return &Uint32{Value: uint32(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'uint32'", args[0].Type())
				}
			},
		},
	},
	{
		"uint64", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16:
					return &Uint64{Value: uint64(arg.(Integer).Integer())}
				case *Uint64:
					return arg
				case *Byte:
					return &Uint64{Value: uint64(arg.Value)}
				case *Rune:
					return &Uint64{Value: uint64(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'uint64'", args[0].Type())
				}
			},
		},
	},
	{
		"byte", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Byte{Value: uint8(arg.(Integer).Integer())}
				case *Uint64:
					return &Byte{Value: uint8(arg.Value)}
				case *Byte:
					return arg
				case *Rune:
					return &Uint8{Value: uint8(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'byte'", args[0].Type())
				}
			},
		},
	},
	{
		"float32", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Float32{Value: float32(arg.(Integer).Integer())}
				case *Uint64:
					return &Float32{Value: float32(arg.Value)}
				case *Float32:
					return arg
				case *Float64:
					return &Float32{Value: float32(arg.Value)}
				case *Byte:
					return &Float32{Value: float32(arg.Value)}
				case *Rune:
					return &Float32{Value: float32(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'float32'", args[0].Type())
				}
			},
		},
	},
	{
		"float64", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *Int, *Int8, *Int16, *Int32, *Int64, *Uint, *Uint8, *Uint16, *Uint32:
					return &Float64{Value: float64(arg.(Integer).Integer())}
				case *Uint64:
					return &Float64{Value: float64(arg.Value)}
				case *Float32:
					return &Float64{Value: float64(arg.Value)}
				case *Float64:
					return arg
				case *Byte:
					return &Float64{Value: float64(arg.Value)}
				case *Rune:
					return &Float64{Value: float64(arg.Value)}
				default:
					return NewError("cannot convert the type '%s' to type 'float64'", args[0].Type())
				}
			},
		},
	},
	// -------------------------------------------
	{
		"len", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. want=1, got=%d", len(args))
				}
				switch arg := args[0].(type) {
				case *String:
					return &Int{Value: len(arg.Value)}
				case *Array:
					return &Int{Value: len(arg.Elements)}
				case *Hash:
					return &Int{Value: len(arg.Pairs)}
				default:
					return NewError("argument to 'len' not support, got %s", args[0].Type())
				}
			},
		},
	},
	{
		"append", 1,
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return NewError("wrong number of arguments. want=2, got=%d", len(args))
				}
				if args[0].Type() != ARRAY_OBJ {
					return NewError("argument to 'append' must be array, got %s", args[0].Type())
				}
				array := args[0].(*Array)

				length := len(array.Elements)
				newElems := make([]Object, length+1, length+1)
				copy(newElems, array.Elements)
				newElems[length] = args[1]
				array.Elements = newElems
				return array
			},
		},
	},
	{
		"println", 0,
		&Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg)
				}
				return nil
			},
		},
	},
}

type builtin struct {
	rtNum int
	fn    *Builtin
}

var builtinsMap = make(map[string]builtin)

func init() {
	for _, item := range Builtins {
		builtinsMap[item.Name] = builtin{rtNum: item.rtName, fn: item.Builtin}
	}
}

func GetBuiltinByName(name string) *Builtin {
	if item, ok := builtinsMap[name]; ok {
		return item.fn
	}
	return nil
}

func GetBuiltinReturnNum(name string) (int, bool) {
	if item, ok := builtinsMap[name]; ok {
		return item.rtNum, ok
	}
	return 0, false
}
