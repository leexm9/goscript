package evaluator

import (
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
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input, true)
		testObject(t, evaluated, tt.expected)
	}
}

func testObject(t *testing.T, evaluated object.Object, expected any) bool {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		return testIntegerObject(t, evaluated, expected)
	case float64:
		return testFloatObject(t, evaluated, expected)
	case bool:
		return testBooleanObject(t, evaluated, expected)
	case byte:
		return testByteObject(t, evaluated, expected)
	case string:
		return testStringObject(t, evaluated, expected)
	case *object.Null:
		return testNullObject(t, evaluated, expected)
	case object.Error:
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", expected, evaluated)
			return false
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%s, got=%s", expected, errObj.Message)
		}
	default:
		t.Errorf("object is not support. got=%T (%+v)", evaluated, evaluated)
		return false
	}
	return true
}

func testIntegerObject(t *testing.T, obj object.Object, expected int) bool {
	t.Helper()
	result, ok := obj.(object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Integer() != int64(expected) {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Integer(), expected)
		return false
	}

	return true
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	t.Helper()
	result, ok := obj.(object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Float() != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Float(), expected)
		return false
	}

	return true
}

func testByteObject(t *testing.T, obj object.Object, expected byte) bool {
	t.Helper()
	result, ok := obj.(*object.Byte)
	if !ok {
		t.Errorf("object is not Byte. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not string. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object, expected any) bool {
	t.Helper()
	if expected == nil && obj == nil {
		return true
	}
	return false
}
