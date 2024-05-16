package compiler

import (
	"fmt"
	"goscript/code"
	"goscript/object"
	"goscript/program"
	"testing"
)

type compilerTestCase struct {
	input             string
	expectedConstants []any
	expectedIns       []code.Instructions
}

func parseProgram(t *testing.T, input string, isStmt bool) *program.Program {
	in := program.Input{Name: "", Content: input, IsStmt: isStmt, IsCheck: false}
	prog, err := program.ParseFile(in)
	if err != nil {
		t.Errorf("parse error: %s", err)
		t.FailNow()
	}

	return prog
}

func runCompilerTests(t *testing.T, tests []compilerTestCase, isStmt bool) {
	t.Helper()

	for _, tt := range tests {
		prog := parseProgram(t, tt.input, isStmt)

		compiler := New()
		err := compiler.CompileProgram(prog)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testConstants(tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}

		err = testInstructions(tt.expectedIns, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
	}
}

func TestObjectArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpADD),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSUB),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMUL),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "2 / 1",
			expectedConstants: []any{2, 1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpQUO),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGTR),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpLSS),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEQL),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNEQ),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEQL),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNEQ),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []any{1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPrefixSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpNOT),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"monkey"`,
			expectedConstants: []any{"monkey"},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []any{"mon", "key"},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpADD),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestVarStatement(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
					var one, two = 1, 2
					one
					`,
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					one := 1
					one
					`,
			expectedConstants: []any{1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					var one = 1
					var two = one
					two
					`,
			expectedConstants: []any{1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]int{}",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[]int{1, 2, 3}",
			expectedConstants: []any{1, 2, 3},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[]int{1 + 2, 3 - 4, 5 * 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpADD),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSUB),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMUL),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "map[string]int{}",
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "map[int]int{1: 2, 3: 4, 5: 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "map[int]int{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpADD),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMUL),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestIndexExpr(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]int{1, 2, 3}[1 + 1]",
			expectedConstants: []any{1, 2, 3, 1, 1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpADD),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "map[int]int{1: 2}[2 - 1]",
			expectedConstants: []any{1, 2, 2, 1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSUB),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					var a = []int{1, 2, 3}
					a[1 + 1]
					`,
			expectedConstants: []any{1, 2, 3, 1, 1},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpADD),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					var a = map[int]int{1: 2, 3: 4 + 5, 6: 7 * 8 }
					a[3]
					`,
			expectedConstants: []any{1, 2, 3, 4, 5, 6, 7, 8, 3},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpADD),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpConstant, 6),
				code.Make(code.OpConstant, 7),
				code.Make(code.OpMUL),
				code.Make(code.OpHash, 6),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 8),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestCompilerScope(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}
	globalSymbolTable := compiler.SymbolTable

	compiler.emit(code.OpMUL)

	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}

	compiler.emit(code.OpSUB)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d", len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpSUB {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpSUB)
	}

	if compiler.SymbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}

	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}

	if compiler.SymbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.SymbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(code.OpADD)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d", len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpADD {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpADD)
	}

	previous := compiler.scopes[compiler.scopeIndex].previousInstruction
	if previous.Opcode != code.OpMUL {
		t.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d", previous.Opcode, code.OpMUL)
	}
}

func TestIfStmt(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `if true { 10 }`,
			expectedConstants: []any{10},
			expectedIns: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 10),
				// 0010
				code.Make(code.OpPop),
			},
		},
		{
			input:             `if true { 10 } else { 20 }`,
			expectedConstants: []any{10, 20},
			expectedIns: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 13),
				// 0010
				code.Make(code.OpConstant, 1),
				// 0013
				code.Make(code.OpPop),
			},
		},
		{
			input:             `if 1 > 2 { 10 }`,
			expectedConstants: []any{1, 2, 10},
			expectedIns: []code.Instructions{
				// 0000
				code.Make(code.OpConstant, 0),
				// 0003
				code.Make(code.OpConstant, 1),
				// 0006
				code.Make(code.OpGTR),
				// 0007
				code.Make(code.OpJumpNotTruthy, 16),
				// 0010
				code.Make(code.OpConstant, 2),
				// 0013
				code.Make(code.OpJump, 16),
				// 0016
				code.Make(code.OpPop),
			},
		},
		{
			input:             `if 1 > 2 { 1 } else if 1 == 2 { 0 } else { -1 }`,
			expectedConstants: []any{1, 2, 1, 1, 2, 0, 1},
			expectedIns: []code.Instructions{
				// 0000
				code.Make(code.OpConstant, 0),
				// 0003
				code.Make(code.OpConstant, 1),
				// 0006
				code.Make(code.OpGTR),
				// 0007
				code.Make(code.OpJumpNotTruthy, 16),
				// 0010
				code.Make(code.OpConstant, 2),
				// 0013
				code.Make(code.OpJump, 36),
				// 0016
				code.Make(code.OpConstant, 3),
				// 0019
				code.Make(code.OpConstant, 4),
				// 0022
				code.Make(code.OpEQL),
				// 0023
				code.Make(code.OpJumpNotTruthy, 32),
				// 0026
				code.Make(code.OpConstant, 5),
				// 0029
				code.Make(code.OpJump, 36),
				// 0032
				code.Make(code.OpConstant, 6),
				// 0035
				code.Make(code.OpPrefixSub),
				// 0036
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests, true)
}

func TestFunction(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
					func() int { return 5 + 10 }
					`,
			expectedConstants: []any{5, 10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpADD),
					code.Make(code.OpReturnValue, 1),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					func() { 5 + 10 }
					`,
			expectedConstants: []any{5, 10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpADD),
					code.Make(code.OpReturn),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					func() { }
					`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestFunctionScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
					func() int {
						var num = 55
						return num
					}
					`,
			expectedConstants: []any{55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue, 1),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					var num = 55
					var b = func() int { return num }
					var a = 3
					var c = b()
					`,
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue, 1),
				},
				3,
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSetGlobal, 2),

				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpCall, 0),
				code.Make(code.OpSetGlobal, 3),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestFunctionCall(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
					func() int { return 24 }()
					`,
			expectedConstants: []any{24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue, 1),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					func(a, b int) int { return a + b }(3, 4)
					`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpADD),
					code.Make(code.OpReturnValue, 1),
				},
				3, 4,
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					noArg := func() int { return 24 }
					noArg()
					`,
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue, 1),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					oneArg = func(a int) { }
					oneArg(24)
					`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
					add = func(a, b int) { return a + b }
					add(3, 4)
					`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpADD),
					code.Make(code.OpReturnValue, 1),
				},
				3, 4,
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestBuiltins(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             ` len([]int{})`,
			expectedConstants: []any{},
			expectedIns: []code.Instructions{
				code.Make(code.OpGetBuiltin, 13),
				code.Make(code.OpArray, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `append([]int{}, 1)`,
			expectedConstants: []any{1},
			expectedIns: []code.Instructions{
				code.Make(code.OpGetBuiltin, 14),
				code.Make(code.OpArray, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			func(a int) func(int) int {
				return func(b int) int { return a + b }
			}
			`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpADD),
					code.Make(code.OpReturnValue, 1),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturnValue, 1),
				},
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
				newClosure := func(a int) func() int {
					return func() int { return a }
				}
				closure := newClosure(99)
				closure()
			`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpReturnValue, 1),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturnValue, 1),
				},
				99,
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 1),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
				var newAdder = func(a, b int) func(int) int {
					return func(c int) { return a + b + c }
				}
				adder := newAdder(1, 2)
				adder(8)
			`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpADD),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpADD),
					code.Make(code.OpReturnValue, 1),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturnValue, 1),
				},
				1, 2, 8,
			},
			expectedIns: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 2),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, true)
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)
	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot=%d", concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction as %d.\nwant=%q\ngot=%q", i, ins, actual[i])
		}
	}
	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func testConstants(expected []any, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. want=%d, got=%d", len(expected), len(actual))
	}
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
		case []code.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", i, actual[i])
			}
			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Integer() != expected {
		return fmt.Errorf("object has wrong value. want=%d, got=%d", expected, result.Integer())
	}
	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%q, got=%q", expected, result.Value)
	}
	return nil
}
