package evaluator

import (
	"fmt"
	"goscript/object"
	"goscript/program"
	"testing"
)

func testEval(t *testing.T, input string, isStmt bool) object.Object {
	in := program.Input{Name: "", Content: input, IsStmt: isStmt, IsCheck: false}
	prog, err := program.ParseFile(in)
	if err != nil {
		t.Errorf("parse error: %s", err)
		t.FailNow()
	}
	return EvalProgram(prog)
}

func TestStmt(t *testing.T) {
	test := struct {
		input    string
		expected any
	}{
		`
				b := 10
				a := []int{1, 2, 3, 4}
				for i := 0; i < len(a); i++ {
					if i%2 == 0 {
						continue
					}
					b = b + a[i]
				}
				b
			`,
		16,
	}
	evaluated := testEval(t, test.input, true)
	testObject(t, evaluated, test.expected)
}

func TestObject(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"5", 5},
		{"3.4", 3.4},
		{`"hello"`, "hello"},
		{`"hello\tworld"`, "hello	world"},
		{`"hello\"world"`, `hello"world`},
		{"'b'", "b"[0]},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestVarStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{
			`
					var a = 5
					a
				`,
			5,
		},
		{
			`
					var a, b = 2, 3
					a
				`,
			2,
		},
		{
			`
					var a, b = 2, 3
					b
				`,
			3,
		},
		{
			`
					a := "str"
					a
				`,
			"str",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestBinaryOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"3 + 4", 7},
		{"3 - 4", -1},
		{"3 * 4", 12},
		{"4 / 2", 2},
		{"4 % 3", 1},
		{"3 & 1", 1},
		{"3 | 1", 3},
		{"3 ^ 1", 2},
		{"3 << 1", 6},
		{"3 >> 1", 1},
		{"3 &^ 1", 2},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},

		{"3 > 1", true},
		{"3 >= 1", true},
		{"3 < 1", false},

		{`"hel" + "lo"`, "hello"},
		{`"hello" + " " + "world"`, "hello world"},
		{`"ab" == "ab"`, true},
		{`"ab" != "ab"`, false},
		{`"ab" != ""`, true},

		{"true == true", true},
		{"true == false", false},
		{"true && false", false},
		{"true || false", true},

		{
			"[]int{1, 2, 4, 5}",
			[]int{1, 2, 4, 5},
		},
		{
			"[]int{1+2, 3-4, 5 * 6}",
			[]int{3, -1, 30},
		},
		{
			`map[string]int{"A":1, "B":2, "C": 3}`,
			map[string]int{"A": 1, "B": 2, "C": 3},
		},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestUnaryOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"-3", -3},
		{"+3", 3},
		{"-(-3)", 3},
		{"-3.2", -3.2},
		{"!true", false},
		{"!false", true},
		{"!!false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestCompositeIndexLit(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"[]int{1, 2, 3, 4}[0]", 1},
		{"[]int{1, 2, 3, 4}[1]", 2},
		{"[]int{1, 2, 3, 4}[1 + 1]", 3},
		{"[]int{1, 2, 3, 4}[3]", 4},
		{"[][]int{[]int{1, 2, 3}}[0][1]", 2},
		{`map[string]int{"A":1, "B":2, "C": 3}["A"]`, 1},
		{`map[string]int{"A":1, "B":2, "C": 3}["B"]`, 2},
		{`map[string]int{"A":1, "B":2, "C": 3}["D"]`, 0},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestCompositeIndexLit2(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{
			`
					var myArray = []int{1, 2, 3}
					myArray[2]
				`,
			3,
		},
		{
			`
					var myArray = []int{1, 2, 3}
					myArray[0] + myArray[1] + myArray[2]
				`,
			6,
		},
		{
			`
					var myArray = []int{1, 2, 3}
					i := myArray[0]
					myArray[i]
				`,
			2,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					m["B"]
				`,
			2,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					m["E"]
				`,
			0,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					b := m["B"]
					b
				`,
			2,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					b, ok := m["B"]
					b
				`,
			2,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					b, ok := m["B"]
					ok
				`,
			true,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					b, ok := m["E"]
					ok
				`,
			false,
		},
		{
			`
					m := map[string]int{
						"A": 1,
						"B": 2,
						"C": 3,
						"D": 4,
					}
					_, ok := m["D"]
					ok
				`,
			true,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"return 10", 10},
		{"return 3 * 2", 6},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"if true { return 10 }", 10},
		{"if false { return 10 }", nil},
		{"if 1 < 2 { return 10 }", 10},
		{"if 1 > 2 { return 10 }", nil},
		{"if 1 < 2 { return 10 } else { return 20 }", 10},
		{"if 1 > 2 { return 10 } else { return 20 }", 20},
		{"if 2 < 2 { return 10 } else if 2 < 3 { return 20 }", 20},
		{"if 3 > 4 { return 1 } else if 3 == 4 { return 0 } else { return -1 }", -1},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		_, ok := tt.expected.(int)
		if ok {
			testObject(t, evaluated, tt.expected)
		} else {
			err := testNullObject(t, evaluated, tt.expected)
			if err != nil {
				t.Errorf("testNullObject failed: %s", err)
			}
		}
	}
}

func TestBuiltinFunc(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`len("")`, 0},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to 'len' not support, got int"},
		{`len("one", "two")`, "wrong number of arguments. want=1, got=2"},
		{`len([]int{1, 2, 3})`, 3},
		{`len([]int{})`, 0},
		{`append([]int{}, 1)`, []int{1}},
		{`append(1, 1)`, "argument to 'append' must be array, got int"},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		_, ok := tt.expected.(string)
		if ok {
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("objcet is not string. got=%T(%+v)", evaluated, evaluated)
				continue
			}
			if errObj.Message != tt.expected {
				t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Message)
			}
		} else {
			testObject(t, evaluated, tt.expected)
		}
	}
}

func TestForExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`
				a := []int{2, 3, 4, 5}
				for i := 0; i < len(a); i++ {
					a[i] = a[i] + i
				}
				a[1]
			`,
			4,
		},
		{`
				a := []int{2, 3, 4, 5}
				b := 0
				for i := 0; i < len(a); i++ {
					b = b + i + a[i]
				}
				b
			`,
			20,
		},
		{`
				a := []int{2, 3, 4, 5}
				var b int
				for _, item := range a {
					b += item
				}
				b
			`,
			14,
		},
		{`
				a := []int{2, 3, 4, 5}
				var b int
				for i := range a {
					b += i
				}
				b
			`,
			6,
		},
		{
			`
					m := map[int]string{
						1: "A",
						2: "B",
						3: "C",
						5: "D",
					}
					b := 0
					for k := range m {
						b += k
					}
					b
				`,
			11,
		},
		{
			`
					m := map[int]string{
						1: "A",
						2: "B",
						3: "C",
						5: "D",
					}
					b := 0
					for _, v := range m {
						b += int(v[0])
					}
					b
				`,
			266,
		},
		{
			`
					m := map[int]string{
						1: "A",
						2: "B",
						3: "C",
						5: "D",
					}
					b := 0
					for k, v := range m {
						b = k + int(v[0]) + b
					}
					b
				`,
			277,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestFunctionLitExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"func(a, b int) int { return a + b }(2, 3)", 5},
		{
			`
					var identity = func(x int) int { return x }
					identity(5)
				`,
			5,
		},
		{
			`
					double := func(x int) int { return x  * 2}
					double(5)
				`,
			10,
		},
		{
			`
					add := func(x, y int) int { return x + y}
					add(3, 5)
				`,
			8,
		},
		{
			`
					add := func(x, y int) int { return x + y}
					add(3 + 5, add(4, 5))
				`,
			17,
		},
		{
			`
					a := func(x int) int {
						return x + 3
					}(5)
					a
				`,
			8,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestClosures(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{
			`
					var newAdder = func(x int) func(int) int {
						return func(y int) int {
							return x + y
						}
					}
					addTwo := newAdder(2)
					a := addTwo(5)
					a
				`,
			7,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{
			`
					package tmp
					
					func main() {
						countDown(1)
					}
					
					func countDown(x int) int {
						if x == 0 {
							return 0
						} else {
							return countDown(x - 1)
						}
					}
				`,
			0,
		},
		{
			`
					package tmp
					
					func main() {
						fibonacci(15)
					}
					
					func fibonacci(x int) int {
						if x == 0 || x == 1 {
							return x
						} else {
							return fibonacci(x - 1) + fibonacci(x - 2)
						}
					}
				`,
			610,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(t, tt.input, false)
		testObject(t, evaluated, tt.expected)
	}
}

func testObject(t *testing.T, evaluated object.Object, expected any) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(t, evaluated, expected)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case float64:
		err := testFloatObject(t, evaluated, expected)
		if err != nil {
			t.Errorf("testFloatObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(t, evaluated, expected)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case byte:
		err := testByteObject(t, evaluated, expected)
		if err != nil {
			t.Errorf("testByteObject failed: %s", err)
		}
	case string:
		err := testStringObject(t, evaluated, expected)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case *object.Null:
		err := testNullObject(t, evaluated, expected)
		if err != nil {
			t.Errorf("testNullObject failed: %s", err)
		}
	case object.Error:
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", expected, evaluated)
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%s, got=%s", expected, errObj.Message)
		}
	case []int:
		array, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("object is not array: %T (%+v)", evaluated, evaluated)
		}
		if len(array.Elements) != len(expected) {
			t.Errorf("wrong number of elements. want=%d, got=%d", len(expected), len(array.Elements))
		}
		for i, elem := range array.Elements {
			err := testIntegerObject(t, elem, expected[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[string]int:
		hashObj, ok := evaluated.(*object.Hash)
		if !ok {
			t.Errorf("object is not hash: %T (%+v)", evaluated, evaluated)
		}
		for _, pair := range hashObj.Pairs {
			key := pair.Key.(*object.String)
			v := pair.Value
			ev, ok1 := expected[key.Value]
			if !ok1 {
				t.Errorf("the key: %s not exist in hash", key.Value)
			}
			err := testIntegerObject(t, v, ev)
			if err != nil {
				t.Errorf("key-value: %s - %d not exist in hash", key.Value, v.(*object.Int).Value)
			}
		}
	default:
		t.Errorf("object is not support. got=%T (%+v)", evaluated, evaluated)
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int) error {
	t.Helper()
	result, ok := obj.(object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
	}
	if result.Integer() != int64(expected) {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Integer(), expected)
	}
	return nil
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) error {
	t.Helper()
	result, ok := obj.(object.Float)
	if !ok {
		return fmt.Errorf("object is not Float. got=%T (%+v)", obj, obj)
	}
	if result.Float() != expected {
		return fmt.Errorf("object has wrong value. got=%f, want=%f", result.Float(), expected)
	}
	return nil
}

func testByteObject(t *testing.T, obj object.Object, expected byte) error {
	t.Helper()
	result, ok := obj.(*object.Byte)
	if !ok {
		return fmt.Errorf("object is not Byte. got=%T (%+v)", obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}
	return nil
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) error {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
	}
	return nil
}

func testStringObject(t *testing.T, obj object.Object, expected string) error {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		return fmt.Errorf("object is not string. got=%T (%+v)", obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
	}
	return nil
}

func testNullObject(t *testing.T, obj object.Object, expected any) error {
	t.Helper()
	if expected == nil && obj == nil {
		return nil
	}
	return fmt.Errorf("object is not nil. got=%T, want=nil", obj)
}
