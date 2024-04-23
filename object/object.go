package object

import "strconv"

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
	BOOLEAN_OBJ
	STRING_OBJ
	NULL_OBJ

	FLOAT32_OBJ
	FLOAT64_OBJ
)

type Object interface {
	Type() ObjectType
	String() string
}

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

type Rune struct {
	Value int32
}

func (r *Rune) Type() ObjectType { return INT32_OBJ }
func (r *Rune) String() string {
	return strconv.FormatInt(int64(r.Value), 10)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) String() string   { return strconv.FormatBool(b.Value) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) String() string   { return s.Value }
