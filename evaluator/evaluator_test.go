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
		`map[string]int{"A":1, "B":2, "C": 3}`,
		map[string]int{"A": 1, "B": 2, "C": 4},
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

		{"3 > 1", true},
		{"3 >= 1", true},
		{"3 < 1", false},

		{`"hel" + "lo"`, "hello"},
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
		{`map[string]int{"A":1, "B":2, "C": 3}["A"]`, 1},
		{`map[string]int{"A":1, "B":2, "C": 3}["B"]`, 2},
		{`map[string]int{"A":1, "B":2, "C": 3}["D"]`, 0},
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
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
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
	return nil
}
