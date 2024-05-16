package object

import (
	"bytes"
	"fmt"
	"go/ast"
	"goscript/code"
	"hash/fnv"
	"strconv"
	"strings"
)

type ObjectType int

const (
	ERROR_OBJ ObjectType = iota
	INT_OBJ
	INT8_OBJ
	INT16_OBJ
	INT32_OBJ
	INT64_OBJ
	UINT_OBJ
	UINT8_OBJ
	UINT16_OBJ
	UINT32_OBJ
	UINT64_OBJ
	FLOAT32_OBJ
	FLOAT64_OBJ
	BOOLEAN_OBJ
	STRING_OBJ
	NULL_OBJ

	ARRAY_OBJ
	HASH_OBJ

	MAP_EXIST_OBJ
	SINGLE_RETURN_OBJ
	MULTI_RETURN_OBJ
	RETURN_OBJ

	CONTINUE_OBJ
	BREAK_OBJ

	FUNCTION_OBJ
	COMPILED_FUNCTION_OBJ
	CLOSURE_OBJ

	FORLOOP_OBJ
	RANGELOOP_OBJ

	BUILTIN_OBJ
)

var typeLiteral = map[ObjectType]string{
	ERROR_OBJ:   "error",
	INT_OBJ:     "int",
	INT8_OBJ:    "int8",
	INT16_OBJ:   "int16",
	INT32_OBJ:   "int32",
	INT64_OBJ:   "int64",
	UINT_OBJ:    "uint",
	UINT8_OBJ:   "uint8",
	UINT16_OBJ:  "uint16",
	UINT32_OBJ:  "uint32",
	UINT64_OBJ:  "uint64",
	FLOAT32_OBJ: "float32",
	FLOAT64_OBJ: "float64",
	BOOLEAN_OBJ: "bool",
	STRING_OBJ:  "string",
	ARRAY_OBJ:   "array",
	HASH_OBJ:    "hash",
}

func (t ObjectType) String() string {
	return typeLiteral[t]
}

func (t ObjectType) IsInteger() bool {
	return ERROR_OBJ < t && t < FLOAT32_OBJ
}

func (t ObjectType) IsFloat() bool {
	return t == FLOAT32_OBJ || t == FLOAT64_OBJ
}

func (t ObjectType) IsRange() bool {
	return t == STRING_OBJ || t == ARRAY_OBJ || t == HASH_OBJ
}

type Object interface {
	Type() ObjectType
	String() string
}

type HashKey struct {
	Type  ObjectType
	Value int64
}

type Hashable interface {
	HashKey() HashKey
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) String() string   { return e.Message }

type Null struct {
}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) String() string   { return "nil" }

type (
	Integer interface {
		Integer() int64
	}

	Int struct {
		Value int
	}

	Int8 struct {
		Value int8
	}

	Int16 struct {
		Value int16
	}

	Int32 struct {
		Value int32
	}

	Int64 struct {
		Value int64
	}

	Uint struct {
		Value uint
	}

	Uint8 struct {
		Value uint8
	}

	Uint16 struct {
		Value uint16
	}

	Uint32 struct {
		Value uint32
	}

	Uint64 struct {
		Value uint64
	}
)

func (i *Int) Integer() int64    { return int64(i.Value) }
func (i *Int8) Integer() int64   { return int64(i.Value) }
func (i *Int16) Integer() int64  { return int64(i.Value) }
func (i *Int32) Integer() int64  { return int64(i.Value) }
func (i *Int64) Integer() int64  { return i.Value }
func (i *Uint) Integer() int64   { return int64(i.Value) }
func (i *Uint8) Integer() int64  { return int64(i.Value) }
func (i *Uint16) Integer() int64 { return int64(i.Value) }
func (i *Uint32) Integer() int64 { return int64(i.Value) }
func (i *Uint64) Integer() int64 { return int64(i.Value) }

func (i *Int) Type() ObjectType    { return INT_OBJ }
func (i *Int8) Type() ObjectType   { return INT8_OBJ }
func (i *Int16) Type() ObjectType  { return INT16_OBJ }
func (i *Int32) Type() ObjectType  { return INT32_OBJ }
func (i *Int64) Type() ObjectType  { return INT64_OBJ }
func (i *Uint) Type() ObjectType   { return UINT_OBJ }
func (i *Uint8) Type() ObjectType  { return UINT8_OBJ }
func (i *Uint16) Type() ObjectType { return UINT16_OBJ }
func (i *Uint32) Type() ObjectType { return UINT32_OBJ }
func (i *Uint64) Type() ObjectType { return UINT64_OBJ }

func (i *Int) String() string    { return strconv.FormatInt(int64(i.Value), 10) }
func (i *Int8) String() string   { return strconv.FormatInt(int64(i.Value), 10) }
func (i *Int16) String() string  { return strconv.FormatInt(int64(i.Value), 10) }
func (i *Int32) String() string  { return strconv.FormatInt(int64(i.Value), 10) }
func (i *Int64) String() string  { return strconv.FormatInt(i.Value, 10) }
func (i *Uint) String() string   { return strconv.FormatUint(uint64(i.Value), 10) }
func (i *Uint8) String() string  { return strconv.FormatUint(uint64(i.Value), 10) }
func (i *Uint16) String() string { return strconv.FormatUint(uint64(i.Value), 10) }
func (i *Uint32) String() string { return strconv.FormatUint(uint64(i.Value), 10) }
func (i *Uint64) String() string { return strconv.FormatUint(i.Value, 10) }

func (i *Int) HashKey() HashKey    { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Int8) HashKey() HashKey   { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Int16) HashKey() HashKey  { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Int32) HashKey() HashKey  { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Int64) HashKey() HashKey  { return HashKey{Type: i.Type(), Value: i.Value} }
func (i *Uint) HashKey() HashKey   { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Uint8) HashKey() HashKey  { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Uint16) HashKey() HashKey { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Uint32) HashKey() HashKey { return HashKey{Type: i.Type(), Value: int64(i.Value)} }
func (i *Uint64) HashKey() HashKey { return HashKey{Type: i.Type(), Value: int64(i.Value)} }

type (
	Float interface {
		Float() float64
	}

	Float32 struct {
		Value float32
	}

	Float64 struct {
		Value float64
	}
)

func (f *Float32) Float() float64 { return float64(f.Value) }
func (f *Float64) Float() float64 { return f.Value }

func (f *Float32) Type() ObjectType { return FLOAT32_OBJ }
func (f *Float64) Type() ObjectType { return FLOAT64_OBJ }

func (f *Float32) String() string {
	return strconv.FormatFloat(float64(f.Value), 'f', -1, 32)
}
func (f *Float64) String() string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

type Byte struct {
	Value uint8
}

func (b *Byte) Type() ObjectType { return UINT8_OBJ }
func (b *Byte) String() string {
	return strconv.FormatUint(uint64(b.Value), 10)
}
func (b *Byte) HashKey() HashKey {
	return HashKey{Type: b.Type(), Value: int64(b.Value)}
}

type Rune struct {
	Value int32
}

func (r *Rune) Type() ObjectType { return INT32_OBJ }
func (r *Rune) String() string {
	return strconv.FormatInt(int64(r.Value), 10)
}
func (r *Rune) HashKey() HashKey {
	return HashKey{Type: r.Type(), Value: int64(r.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) String() string   { return strconv.FormatBool(b.Value) }
func (b *Boolean) HashKey() HashKey {
	if b.Value {
		return HashKey{Type: b.Type(), Value: 1}
	} else {
		return HashKey{Type: b.Type(), Value: 0}
	}
}

type (
	String struct {
		Value string
	}

	Array struct {
		ElemType ObjectType
		Elements []Object
		Len      int
	}

	HashPair struct {
		Key   Object
		Value Object
	}

	Hash struct {
		KeyType   ObjectType
		ValueType ObjectType
		Pairs     map[HashKey]HashPair
	}
)

func NewArray(elemType ObjectType, elems []Object, fixed bool, fixLen int) Array {
	if fixed && len(elems) == 0 && fixLen > 0 {
		for i := 0; i < fixLen; i++ {
			elems = append(elems, GetDefaultObject(elemType.String()))
		}
	}
	array := Array{elemType, elems, len(elems)}
	return array
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: int64(h.Sum64())}
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (a *Array) Type() ObjectType  { return ARRAY_OBJ }
func (h *Hash) Type() ObjectType   { return HASH_OBJ }

func (s *String) String() string { return s.Value }
func (a *Array) String() string {
	var out bytes.Buffer
	var elems []string
	for _, element := range a.Elements {
		elems = append(elems, element.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elems, ", "))
	out.WriteString("]")
	return out.String()
}
func (h *Hash) String() string {
	var out bytes.Buffer
	var pairs []string
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s :%s", pair.Key.String(), pair.Value.String()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

/*-----------------------------------*/

type MapExist struct {
	Value     Object
	Exist     bool
	SkipValue bool
}

func (me *MapExist) Type() ObjectType { return MAP_EXIST_OBJ }
func (me *MapExist) String() string   { return me.Value.String() }

type SingleReturn struct {
	Value   Object
	FromFun bool
}

func (srt *SingleReturn) Type() ObjectType { return SINGLE_RETURN_OBJ }
func (srt *SingleReturn) String() string   { return srt.Value.String() }

type MultiReturn struct {
	Values  []Object
	FromFun bool
}

func (mrt *MultiReturn) Type() ObjectType { return MULTI_RETURN_OBJ }
func (mrt MultiReturn) String() string {
	var rts []string
	for _, value := range mrt.Values {
		rts = append(rts, value.String())
	}
	return strings.Join(rts, " ")
}

type Return struct {
}

func (rt *Return) Type() ObjectType { return RETURN_OBJ }
func (rt *Return) String() string   { return "return" }

type Continue struct {
}

func (cn *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (cn *Continue) String() string   { return "continue" }

type Break struct {
}

func (bk *Break) Type() ObjectType { return BREAK_OBJ }
func (bk *Break) String() string   { return "break" }

/*-----------------------------------*/

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (bti Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (bti Builtin) String() string   { return "Builtin function" }

type ElemTypeEnum int

const (
	ElemBase ElemTypeEnum = iota
	ElemArray
	ElemHash
	ElemStruct
)

type (
	ElemType struct {
		Type     *ast.Ident
		TypeElem ElemTypeEnum
		Types    []*ast.Ident
	}

	FunArg struct {
		Symbol *ast.Ident
		Type   ElemType
	}

	FunResult struct {
		Symbol  *ast.Ident
		Type    ElemType
		IsFun   bool
		Params  []FunArg
		Results []FunResult
	}

	Function struct {
		Name    string
		Params  []FunArg
		Results []FunResult
		Body    *ast.BlockStmt
		Env     *Environment
	}
)

func (fn *Function) Type() ObjectType { return FUNCTION_OBJ }
func (fn *Function) String() string   { return "function" }

type CompiledFunction struct {
	Name         string
	Instructions code.Instructions
	NumLocals    int
	NumParams    int
	NumResult    int
	FreeNum      int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) String() string   { return fmt.Sprintf("CompiledFunction[%p]", cf) }

type Closure struct {
	Fn      *CompiledFunction
	ForLoop Object
	Free    []Object
}

func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }
func (c *Closure) String() string   { return fmt.Sprintf("Closure[%p]", c) }

type ForLoop struct {
	Init      code.Instructions
	Cond      code.Instructions
	Body      code.Instructions
	Post      code.Instructions
	NumLocals int
	FreeNum   int
}

func (fl *ForLoop) Type() ObjectType { return FORLOOP_OBJ }
func (fl *ForLoop) String() string   { return fmt.Sprintf("fori") }

type RangeLoop struct {
	X           code.Instructions
	Body        code.Instructions
	IsAnonymous bool
	NumLocals   int
	FreeNum     int
}

func (rl *RangeLoop) Type() ObjectType { return RANGELOOP_OBJ }
func (rl *RangeLoop) String() string   { return fmt.Sprintf("forr") }
