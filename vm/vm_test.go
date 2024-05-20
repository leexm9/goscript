package vm

import (
	"fmt"
	"goscript/compiler"
	"goscript/object"
	"goscript/program"
	"testing"
)

type vmTestCase struct {
	input    string
	expected any
}

func runVmTests(t *testing.T, tests []vmTestCase, isStmt bool) {
	t.Helper()

	for _, tt := range tests {
		prog := parseProgram(t, tt.input, isStmt)
		comp := compiler.New()
		err := comp.CompileProgram(prog)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func parseProgram(t *testing.T, input string, isStmt bool) *program.Program {
	in := program.Input{Name: "", Content: input, IsStmt: isStmt, IsCheck: false}
	prog, err := program.ParseFile(in)
	if err != nil {
		t.Errorf("parser error: %s", err)
		t.FailNow()
	}
	return prog
}

func TestTest(t *testing.T) {
	tt := []vmTestCase{
		{
			`
					package tmp
					
					func main() {
						b := 0
						a := []int{1, 2, 3, 4, 5}
						for i := 0; i < len(a); i++ {
							if i == 0 {
								a[i] = a[i] + 15
							}
							if i < 1 {
								continue
							}
							if i > 3 {
								break
							}
							b += a[i]
						}
						println(a[0])
						println(b)
						println("-----------------")
					
						mm := map[string]int{"A":1, "B":2}
						m, ok := mm["A"]
						println(m)
						println(ok)
						_, ok = mm["C"]
						println(ok)
						println("-----------------")
					
						e, f := ff(3, 4)
						println(e)
						println(f)
						println("-----------------")
					
						var f1 = func(a, b int) int {
							return a + b + 10
						}
						f = f1(f, 4)
						println(f)
						println("-----------------")
					
						println(fibonacci(15))
						println("-----------------")
					
						var adder = func(x int) func(int) int {
							return func(a int) int { return a + x }
						}
						addTwo := adder(2)
						println(addTwo(5))
						println("-----------------")
					}
					
					func add(a, b int) int {
						return a + b
					}
					
					func ff(a, b int) (int, int){
						c := add(a, b)
						return c, a - b
					}
					
					func fibonacci(x int) int {
						if x == 0 || x == 1 {
							return x
						} else {
							return fibonacci(x-1) + fibonacci(x-2)
						}
					}
    			`,
			9,
		},
	}
	runVmTests(t, tt, false)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"-5", -5},
		{"-50 + 100 + -50", 0},
		{"+50 + 100 + -50", 100},
		{"5 * (2 + 10)", 60},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests, true)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
	}

	runVmTests(t, tests, true)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if true { return 10 }", 10},
		{"if true { return 10 } else { return 20 }", 10},
		{"if false { return 10 } else { return 20 }", 20},
		{"if 1 < 2 { return 10 }", 10},
		{"if 1 < 2 { return 10 } else { return 20 }", 10},
		{"if 1 > 2 { return 10 } else { return 20 }", 20},
	}

	runVmTests(t, tests, true)
}

func TestGlobalLetStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			`
				 var one = 1
				 one
				`,
			1,
		},
		{
			`
				 var one = 1
				 var two = 2
				 one + two
				`,
			3,
		},
		{
			`
				 var one = 1
				 var two = one + one
				 one + two
				`,
			3,
		},
	}

	runVmTests(t, tests, true)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVmTests(t, tests, true)
}

func TestArrayLiteral(t *testing.T) {
	tests := []vmTestCase{
		{"[]int{}", []int{}},
		{"[]int{1, 2, 3}", []int{1, 2, 3}},
		{"[]int{1 + 2, 3 - 4, 5 * 6}", []int{3, -1, 30}},
	}

	runVmTests(t, tests, true)
}

func TestHashLiteral(t *testing.T) {
	tests := []vmTestCase{
		{"map[int]int{1:2, 2:3}", map[object.HashKey]int{
			(&object.Int{Value: 1}).HashKey(): 2,
			(&object.Int{Value: 2}).HashKey(): 3,
		}},
		{`map[string]int{"A": 2 * 2, "B": 4 * 4}`, map[object.HashKey]int{
			(&object.String{Value: "A"}).HashKey(): 4,
			(&object.String{Value: "B"}).HashKey(): 16,
		}},
	}

	runVmTests(t, tests, true)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[]int{1, 2, 3}[1]", 2},
		{"[]int{1, 2, 3}[0 + 2]", 3},
		{"[][]int{[]int{1, 2, 3}}[0][0]", 1},
		{
			`
				 var a = []int{1, 2 + 3, 4, 5}
				 a[0]
				`,
			1,
		},
		{"map[int]int{1: 2, 2: 3}[1]", 2},
		{`map[int]string{1: "A", 2: "B"}[2]`, "B"},
		{
			`
				 a := map[string]int{"A": 2, "B": 2 + 3, "C": 5 * 2}
				 a["B"]
				`,
			5,
		},
		{
			`
				 a := map[string]int{"A": 2, "B": 2 + 3, "C": 5 * 2}
				 b, ok = a["C"]
				 b
				`,
			10,
		},
		{
			`
				 a := map[string]int{"A": 2, "B": 2 + 3, "C": 5 * 2}
				 b, ok = a["C"]
				 ok
				`,
			true,
		},
		{
			`
				 a := map[string]int{"A": 2, "B": 2 + 3, "C": 5 * 2}
				 _, ok = a["C"]
				 ok
				`,
			true,
		},
	}

	runVmTests(t, tests, true)
}

func TestIfStmt(t *testing.T) {
	tests := []vmTestCase{
		{
			`
					b, t := 0, 0
					ok := false
					mm := map[string]int{"A":11, "B":12}
					if t, ok = mm["A"]; ok {
						b += t
					}
					ok
				`,
			true,
		},
		{
			`
					b, t := 0, 0
					ok := false
					mm := map[string]int{"A":11, "B":12}
					if t, ok := mm["A"]; ok {
						b += t
					}
					ok
				`,
			false,
		},
		{
			`
					i, b := 0, 0
					a := []int{1, 2, 3, 4}
					for i = 1; i < len(a); i++ {
						b += a[i]
					}
					i
				`,
			4,
		},
		{
			`
					i, b := 0, 0
					a := []int{1, 2, 3, 4}
					for i := 0; i < len(a); i++ {
						b += a[i]
					}
					i
				`,
			0,
		},
	}

	runVmTests(t, tests, true)
}

func TestForAndRangeExpression(t *testing.T) {
	tests := []vmTestCase{
		{
			`
					a := []int{1, 2, 3, 4, 5}
					for i := 0; i < len(a); i++ {
						a[i] = a[i] + i
					}
					a[1]
				`,
			3,
		},
		{
			`
					a := []int{1, 2, 3, 4, 5}
					var b int
					for i := 0; i < len(a); i++ {
						b = b + i + a[i]
					}
					b
				`,
			25,
		},
		{
			`
					a := []int{1, 2, 3, 4, 5}
					var b int
					for _, item := range a {
						b += item
					}
					b
				`,
			15,
		},
		{
			`
					a := []int{1, 2, 3, 4, 5}
					var b int
					for i := range a {
						b += i
					}
					b
				`,
			10,
		},
		{
			`
					a := []int{1, 2, 3, 4, 5}
					var b int
					for i, item := range a {
						b = i + item + b
					}
					b
				`,
			25,
		},
		{
			`
					a := []int{1, 2, 3, 4, 5}
					var b int
					for i := 0; i < len(a); i++ {
						if i < 1 {
							continue
						} else if i > 3 {
							break
						}
						b += a[i]
					}
					b
   				`,
			9,
		},
		{
			`
					m := map[int]string{1: "A", 2: "B", 3: "C", 5: "D"}
					var b int
					for k := range m {
						b += k
					}
					b
				`,
			11,
		},
		{
			`
					m := map[int]string{1: "A", 2: "B", 3: "C", 5: "D"}
					var b int
					for k := range m {
						if k == 3 {
							continue
						}
						b += k
					}
					b
				`,
			8,
		},
		{
			`
					m := map[int]string{1: "A", 2: "B", 3: "C", 5: "D"}
					var b int
					for _, v := range m {
						b += int(v[0])
					}
					b
				`,
			266,
		},
		{
			`
					m := map[int]string{1: "A", 2: "B", 3: "C", 5: "D"}
					var b int
					for k, v := range m {
						b = k + int(v[0]) + b
					}
					b
				`,
			277,
		},
	}

	runVmTests(t, tests, true)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			`
					fivePlusTen := func() int { return 5 + 10 }
					fivePlusTen()
				`,
			15,
		},
		{
			`
					var one = func() int { return 1 }
					var two = func() int { return 2 }
					one() + two()
				`,
			3,
		},
		{
			`
					a := func() int { return 1 }
					b := func() int { return a() + 1 }
					var c = func() int { return b() + 1 }
					c()
				`,
			3,
		},
	}

	runVmTests(t, tests, true)
}

func TestCallingFunctionsWithArgumentAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			`
					one := func() int { 
						one := 1
						return one 
					}
					one();
    			`,
			1,
		},
		{
			`
					oneAndTwo := func() int { 
						var one, two = 1, 2
						return one + two 
					}
					oneAndTwo();
    			`,
			3,
		},
		{
			`
					oneAndTwo := func() int { 
						var one, two = 1, 2
						return one + two 
					}
					threeAndFour := func() int { 
						var three, four = 3, 4
						return three + four 
					}
					oneAndTwo() + threeAndFour();
				`,
			10,
		},
		{
			`
					var global = 50
					minusOne := func() {
						var num = 1
						global = global - num
					}
					var minusTwo = func() {
						num := 2
						global -= num
					}
					minusOne()
					minusTwo()
					global
				`,
			47,
		},
		{
			`
					identity := func(a int) int { return a}
					identity(4)
				`,
			4,
		},
		{
			`
					sum := func(a, b int) int { return a + b }
					sum(4, 5)
				`,
			9,
		},
		{
			`
					sum := func(a, b int) int {
						c := a + b
						return c
					}
					sum(1, 2)
				`,
			3,
		},
		{
			`
					sum := func(a, b int) int {
						c := a + b
						return c
					}
					sum(1, 2) + sum(3, 4)
				`,
			10,
		},
		{
			`
					sum := func(a, b int) int {
						c := a + b
						return c
					}
					outer := func() int {
						return sum(1, 2) + sum(3, 4)
					}
					outer()
				`,
			10,
		},
		{
			`
					globalNum := 10
					sum := func(a, b int) int {
						c := a + b
						c += globalNum
						return c
					}
					outer := func() int {
						return sum(1, 2) + sum(3, 4) + globalNum
					}
					outer() + globalNum
				`,
			50,
		},
	}

	runVmTests(t, tests, true)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			`
					func() { 1 }(1)
				`,
			"execute function wrong number of arguments: want=0, got=1",
		},
		{
			`
					func(a int) { a }()
				`,
			"execute function wrong number of arguments: want=1, got=0",
		},
		{
			`
					func(a, b int) { a + b }(1)
				`,
			"execute function wrong number of arguments: want=2, got=1",
		},
	}

	for _, tt := range tests {
		prog := parseProgram(t, tt.input, true)
		comp := compiler.New()
		err := comp.CompileProgram(prog)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected vm error but resulted in none.")
		}

		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err.Error())
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{Message: "argument to 'len' not support, got int"},
		},
		{
			`len("one", "two")`,
			&object.Error{Message: "wrong number of arguments. want=1, got=2"},
		},
		{`len([]int{1, 2, 3})`, 3},
		{`len([]int{})`, 0},
		{`len(map[string]int{"A": 1, "B": 2})`, 2},
		{
			`
					a := []int{1}
					a = append(a, 2)
					a
				`,
			[]int{1, 2},
		},
		{
			`append(1, 1)`,
			&object.Error{Message: "argument to 'append' must be array, got int"},
		},
	}

	runVmTests(t, tests, true)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			`
					returnsOne := func(a int) int { return 1 + a }
					returnsOneReturner := func() func(int) int { return returnsOne }
					returnsOneReturner()(3)
				`,
			4,
		},
		{
			`
					returnOneReturner := func() func() int {
					returnOne := func() int { return 1 }
					return returnOne
					}
					returnOneReturner()()
				`,
			1,
		},
		{
			`
					var newClosure = func(a int) func() int {
						return func() int {
							return a
						}
					}
					closure := newClosure(99)
					closure()
				`,
			99,
		},
		{
			`
					newAdder := func(a, b int) func(int) int {
						return func(c int) int {
							return a + b + c
						}
					}
					adder := newAdder(1, 2)
					adder(8)
				`,
			11,
		},
		{
			`
					newAdder := func(a, b int) func(int) int {
						c := a + b
						return func(d int) int {
							return c + d
						}
					}
					adder := newAdder(1, 2)
					adder(8)
				`,
			11,
		},
		{
			`
					newAdderOuter := func(a, b int) func(int) func(int) int {
						c := a + b
						return func(d int) func(int) int {
							e := d + c
							return func(f int) int { return e + f }
						}
					}
					newAdderInner := newAdderOuter(1, 2)
					adder := newAdderInner(3)
					adder(8)
				`,
			14,
		},
		{
			`
					a := 1
					newAdderOuter := func(b int) func(int) func(int) int {
						return func(c int) func(int) int {
							return func(d int) int { return a + b + c + d }
						}
					}
					newAdderInner := newAdderOuter(2)
					adder := newAdderInner(3)
					adder(8)
				`,
			14,
		},
		{
			`
					newClosure := func(a, b int) func() int {
						one := func() int { return a }
						two := func() int { return b }
						return func() int { return one() + two() }
					}
					closure := newClosure(9, 90)
					closure()
				`,
			99,
		},
	}

	runVmTests(t, tests, true)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
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
							return fibonacci(x-1) + fibonacci(x-2)
						}
					}
				`,
			610,
		},
	}

	runVmTests(t, tests, false)
}

func testExpectedObject(t *testing.T, expected any, actual object.Object) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(t, expected, actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case float64:
		err := testFloatObject(t, expected, actual)
		if err != nil {
			t.Errorf("testFloatObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(t, expected, actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case byte:
		err := testByteObject(t, expected, actual)
		if err != nil {
			t.Errorf("testByteObject failed: %s", err)
		}
	case string:
		err := testStringObject(t, expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case *object.Null:
		err := testNullObject(t, expected, actual)
		if err != nil {
			t.Errorf("testNullObject failed: %s", err)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object is not array: %T (%+v)", actual, actual)
		}
		if len(array.Elements) != len(expected) {
			t.Errorf("wrong number of elements. want=%d, got=%d", len(expected), len(array.Elements))
		}
		for i, elem := range array.Elements {
			err := testIntegerObject(t, expected[i], elem)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[object.HashKey]int:
		hashObj, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is not hash: %T (%+v)", actual, actual)
		}
		if len(hashObj.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(hashObj.Pairs))
			return
		}
		for ekey, eValue := range expected {
			pair, ok := hashObj.Pairs[ekey]
			if !ok {
				t.Errorf("no pair for given key in pairs")
			}
			err := testIntegerObject(t, eValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case *object.Error:
		errObj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", expected, actual)
			return
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%s, got=%s", expected, errObj.Message)
		}
	default:
		t.Errorf("object is not support. got=%T (%+v)", actual, actual)
	}
}

func testIntegerObject(t *testing.T, expected int, obj object.Object) error {
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

func testFloatObject(t *testing.T, expected float64, obj object.Object) error {
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

func testByteObject(t *testing.T, expected byte, obj object.Object) error {
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

func testBooleanObject(t *testing.T, expected bool, obj object.Object) error {
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

func testStringObject(t *testing.T, expected string, obj object.Object) error {
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

func testNullObject(t *testing.T, expected any, obj object.Object) error {
	t.Helper()
	if expected == nil && obj == nil {
		return nil
	}
	return fmt.Errorf("object is not nil. got=%T, want=nil", obj)
}
