package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"goscript/code"
	"goscript/object"
	"goscript/program"
	"strconv"
	"strings"
)

var tokenFile *token.File

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions code.Instructions

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Constants    []object.Object
	SymbolTable  *SymbolTable
	Instructions code.Instructions
	GlobalDecls  int
}

type Compiler struct {
	constants   []object.Object
	SymbolTable *SymbolTable

	scopes      []CompilationScope
	scopeIndex  int
	globalDecls int
}

func New() *Compiler {
	global := NewSymbolTable()
	for i, item := range object.Builtins {
		global.DefineBuiltin(i, item.Name)
	}

	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   []object.Object{},
		SymbolTable: global,
		scopes:      []CompilationScope{mainScope},
	}
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
		SymbolTable:  c.SymbolTable,
		GlobalDecls:  c.globalDecls,
	}
}

func (c *Compiler) addConstants(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	newInsPos := len(c.currentInstructions())
	newInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = newInstructions
	return newInsPos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	c.scopes[c.scopeIndex].instructions = old[:last.Position]
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions: code.Instructions{},

		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.SymbolTable = NewEnclosedSymbolTable(c.SymbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.SymbolTable = c.SymbolTable.Outer
	return instructions
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}

func (c *Compiler) storeSymbol(s Symbol) {
	if s.Name == "_" {
		c.emit(code.OpSetNil)
	} else if s.Scope == GlobalScope {
		c.emit(code.OpSetGlobal, s.Index)
	} else if s.Scope == LocalScope {
		c.emit(code.OpSetLocal, s.Index)
	} else if s.Scope == FreeScope {
		c.emit(code.OpSetFree, s.Index)
	} else {
		panic(fmt.Sprintf("not known symbol scope: %s", s.Scope))
	}
}

func (c *Compiler) compileAstFunction(fnName string, fn *object.Function, symbolTable *SymbolTable, constantNum int) error {
	fnTy := &object.Function{}
	fnTy.Params = fn.Params
	fnTy.Results = fn.Results

	_ = symbolTable.DefineWithType(fnName, fnTy)
	c.constants = append(c.constants, nil)

	compiledFn, err := c.compileFunction(fn, fnName)
	if err != nil {
		return err
	}

	compiledFn.Name = fnName
	c.constants[constantNum] = compiledFn
	return nil
}

func (c *Compiler) compile(node ast.Node, defaultType object.Object) error {
	switch node := node.(type) {
	case *ast.DeclStmt:
		return c.compile(node.Decl, defaultType)
	case *ast.GenDecl:
		return c.compileGenDecl(node)
	case *ast.ExprStmt:
		return c.compile(node.X, defaultType)
	case *ast.AssignStmt:
		return c.compileAssignStmt(node)
	case *ast.IncDecStmt:
		return c.compileIncDecStmt(node)
	case *ast.ReturnStmt:
		return c.compileReturnStmt(node)
	case *ast.BlockStmt:
		return c.compileBlockStmt(node, defaultType)
	case *ast.IfStmt:
		return c.compileIfStmt(node)
	case *ast.ForStmt:
		return c.compileForStmt(node)
	case *ast.RangeStmt:
		return c.compileRangeStmt(node)
	case *ast.BinaryExpr:
		return c.compileBinaryExpr(node)
	case *ast.ParenExpr:
		return c.compile(node.X, defaultType)
	case *ast.UnaryExpr:
		err := c.compile(node.X, defaultType)
		if err != nil {
			return err
		}
		switch node.Op {
		case token.NOT:
			c.emit(code.OpNOT)
		case token.SUB:
			c.emit(code.OpPrefixSub)
		case token.ADD:
		default:
			return fmt.Errorf("operator %s not support", node.Op)
		}
	case *ast.IndexExpr:
		_, err := c.compileIndexExpr(node)
		if err != nil {
			return err
		}
	case *ast.FuncLit:
		_, err := c.compileFuncLit(node)
		if err != nil {
			return err
		}
	case *ast.CallExpr:
		return c.compileCallExpr(node)
	case *ast.CompositeLit:
		_, err := c.compileCompositeLit(node, defaultType)
		if err != nil {
			return err
		}
	case *ast.Ident:
		_, err := c.compileIdent(node)
		if err != nil {
			return err
		}
	case *ast.BasicLit:
		_, err := c.compileBasicLit(node, defaultType)
		if err != nil {
			return err
		}
	case *ast.BranchStmt:
		if node.Tok == token.CONTINUE {
			c.emit(code.OpContinue)
			return nil
		} else if node.Tok == token.BREAK {
			c.emit(code.OpBreak)
			return nil
		} else {
			panic(fmt.Sprintf("compiler: not support ast.BranchStmt %s", node.Tok))
		}
	default:
		panic(fmt.Sprintf("compiler: not support ast type %T", node))
	}
	return nil
}

func (c *Compiler) compileBasicLit(node *ast.BasicLit, defaultType object.Object) (object.Object, error) {
	basic, err := parseBasicLit(node)
	if err != nil {
		return nil, err
	}
	if defaultType != nil && basic.Type() != defaultType.Type() {
		basic = object.ConvertValueWithType(basic, defaultType)
		if object.IsError(basic) {
			return nil, errors.New(basic.(*object.Error).Message)
		}
	}
	c.emit(code.OpConstant, c.addConstants(basic))
	return basic, nil
}

func (c *Compiler) compileGenDecl(node *ast.GenDecl) error {
	line, column := parsePos(node.Pos())
	switch node.Tok {
	case token.CONST:
		return fmt.Errorf("%d:%d not support const", line, column)
	case token.VAR:
		for _, spec := range node.Specs {
			err := c.compileGenVar(spec.(*ast.ValueSpec))
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("%d:%d ast.GenDecl with not known type: %s", line, column, node.Tok)
	}
}

func (c *Compiler) compileAssignStmt(node *ast.AssignStmt) error {
	switch node.Tok {
	case token.DEFINE, token.ASSIGN:
		return c.compileDefineStmt(node)
	case token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN,
		token.AND_ASSIGN, token.OR_ASSIGN, token.XOR_ASSIGN, token.SHL_ASSIGN, token.SHR_ASSIGN, token.AND_NOT_ASSIGN:
		return c.compileBinaryAssign(node)
	}
	return nil
}

func (c *Compiler) compileGenVar(spec *ast.ValueSpec) error {
	var vars []string
	for _, ts := range spec.Names {
		vars = append(vars, ts.Name)
	}

	var defObj object.Object
	if spec.Type != nil {
		defObj = object.GetDefaultValueWithExpr(spec.Type)
	}

	if spec.Values == nil {
		idx := c.addConstants(defObj)
		for _, v := range vars {
			c.emit(code.OpConstant, idx)
			symbol := c.SymbolTable.DefineWithType(v, defObj)
			c.storeSymbol(symbol)
		}
		return nil
	}

	n := 0
	for _, expr := range spec.Values {
		switch expr := expr.(type) {
		case *ast.BasicLit:
			obj, err := c.compileBasicLit(expr, defObj)
			if err != nil {
				return err
			}
			symbol := c.SymbolTable.DefineWithType(vars[n], obj)
			c.storeSymbol(symbol)
			n++
		case *ast.Ident:
			symbol, err := c.compileIdent(expr)
			if err != nil {
				return err
			}
			symbol = c.SymbolTable.DefineWithType(vars[n], symbol.Type)
			c.storeSymbol(symbol)
			n++
		case *ast.CallExpr:
			fun := expr.Fun
			switch fn := fun.(type) {
			case *ast.Ident:
				symbol, ok := c.SymbolTable.Resolve(fn.Name)
				if !ok {
					return fmt.Errorf("undefined: %s", fn.Name)
				}
				c.loadSymbol(symbol)

				switch symbol.Type.Type() {
				case object.FUNCTION_OBJ:
					fnc := symbol.Type.(*object.Function)
					for _, arg := range expr.Args {
						err := c.compile(arg, nil)
						if err != nil {
							return err
						}
					}
					nArgs := len(fnc.Params)
					nRet := 0
					for _, result := range fnc.Results {
						if result.Symbol != nil {
							obj := object.GetDefaultValueFromElem(result.Type)
							c.emit(code.OpConstant, c.addConstants(obj))
							nRet++
						}
					}
					c.emit(code.OpCall, nArgs+nRet)

					if fnc.Results[0].IsFun {
						rtSymbol := Symbol{Type: &object.Function{Params: fnc.Results[0].Params, Results: fnc.Results[0].Results}}
						err := c.assignValue(Variable{Name: vars[0], Type: VarIdent}, &rtSymbol)
						if err != nil {
							return err
						}
						n++
					} else {
						for i, s := range vars {
							rtSymbol := Symbol{Type: object.GetDefaultValueFromElem(fnc.Results[i].Type)}
							err := c.assignValue(Variable{Name: s, Type: VarIdent}, &rtSymbol)
							if err != nil {
								return err
							}
							n++
						}
					}
				case object.BUILTIN_OBJ:
					for _, arg := range expr.Args {
						err := c.compile(arg, nil)
						if err != nil {
							return err
						}
					}
					c.emit(code.OpCall, len(expr.Args))

					err := c.assignValue(Variable{Name: vars[0], Type: VarIdent}, nil)
					if err != nil {
						return err
					}
					n++
				}
			}
		case *ast.IndexExpr:
			rtSymbol, err := c.compileIndexExpr(expr)
			if err != nil {
				return err
			}

			if rtSymbol.Type.Type() == object.ARRAY_OBJ {
				if defObj == nil {
					defObj = object.GetDefaultObject(rtSymbol.Type.(*object.Array).ElemType.String())
				}
				symbol := c.SymbolTable.DefineWithType(vars[n], defObj)
				c.storeSymbol(symbol)
				n++
			} else if rtSymbol.Type.Type() == object.HASH_OBJ {
				if defObj == nil {
					defObj = object.GetDefaultObject(rtSymbol.Type.(*object.Hash).ValueType.String())
				}
				symbol := c.SymbolTable.DefineWithType(vars[n], defObj)
				c.storeSymbol(symbol)
				n++

				if len(vars) == 2 && len(spec.Values) == 1 {
					defObj = object.GetDefaultObject(object.BOOLEAN_OBJ.String())
					symbol = c.SymbolTable.DefineWithType(vars[n], defObj)
					c.storeSymbol(symbol)
					n++
				}
			}
		case *ast.BinaryExpr:
			err := c.compileBinaryExpr(expr)
			if err != nil {
				return err
			}
			symbol := c.SymbolTable.Define(vars[n])
			c.storeSymbol(symbol)
			n++
		case *ast.FuncLit:
			fnSymbol, err := c.compileFuncLit(expr)
			if err != nil {
				return err
			}
			symbol := c.SymbolTable.DefineWithType(vars[n], fnSymbol.Type)
			c.storeSymbol(symbol)
			n++
		case *ast.CompositeLit:
			rtSymbol, err := c.compileCompositeLit(expr, defObj)
			if err != nil {
				return err
			}
			symbol := c.SymbolTable.DefineWithType(vars[n], (*rtSymbol).Type)
			c.storeSymbol(symbol)
			n++
		default:
			line, column := parsePos(expr.Pos())
			return fmt.Errorf("%d:%d GenVar not support %T", line, column, expr)
		}
	}

	if len(vars) != n {
		tn := spec.Names[n]
		tl, tc := parsePos(tn.Pos())
		return fmt.Errorf("%d:%d missing init expr for '%s'", tl, tc, tn.Name)
	}
	return nil
}

func (c *Compiler) compileDefineStmt(node *ast.AssignStmt) error {
	var vars []Variable
	for i := 0; i < len(node.Lhs); i++ {
		switch item := node.Lhs[i].(type) {
		case *ast.Ident:
			vars = append(vars, Variable{Name: item.Name, Type: VarIdent})
		case *ast.IndexExpr:
			tmp, _ := item.X.(*ast.Ident)
			variable := Variable{Name: tmp.Name, Type: VarIdent}
			variable.Index = item.Index
			vars = append(vars, variable)
		default:
			line, column := parsePos(item.Pos())
			return fmt.Errorf("%d:%d not support", line, column)
		}
	}

	n := 0
	for _, expr := range node.Rhs {
		switch expr := expr.(type) {
		case *ast.BasicLit:
			basic, err := parseBasicLit(expr)
			if err != nil {
				return err
			}
			variable := vars[n]
			c.emit(code.OpConstant, c.addConstants(basic))
			err = c.assignValue(variable, &Symbol{Type: basic})
			if err != nil {
				return err
			}
			n++
		case *ast.Ident:
			symbol, err := c.compileIdent(expr)
			if err != nil {
				return err
			}
			variable := vars[n]
			err = c.assignValue(variable, &symbol)
			if err != nil {
				return err
			}
			n++
		case *ast.CallExpr:
			fun := expr.Fun
			switch fn := fun.(type) {
			case *ast.Ident:
				symbol, ok := c.SymbolTable.Resolve(fn.Name)
				if !ok {
					return fmt.Errorf("undefined: %s", fn.Name)
				}
				c.loadSymbol(symbol)

				switch symbol.Type.Type() {
				case object.FUNCTION_OBJ:
					fnc := symbol.Type.(*object.Function)
					for _, arg := range expr.Args {
						err := c.compile(arg, nil)
						if err != nil {
							return err
						}
					}
					nArgs := len(fnc.Params)
					nRet := 0
					for _, result := range fnc.Results {
						if result.Symbol != nil {
							obj := object.GetDefaultValueFromElem(result.Type)
							c.emit(code.OpConstant, c.addConstants(obj))
							nRet++
						}
					}
					c.emit(code.OpCall, nArgs+nRet)

					if fnc.Results[0].IsFun {
						rtSymbol := Symbol{Type: &object.Function{Params: fnc.Results[0].Params, Results: fnc.Results[0].Results}}
						err := c.assignValue(vars[0], &rtSymbol)
						if err != nil {
							return err
						}
					} else {
						for i := 0; i < len(vars); i++ {
							rtSymbol := Symbol{Type: object.GetDefaultValueFromElem(fnc.Results[i].Type)}
							err := c.assignValue(vars[0], &rtSymbol)
							if err != nil {
								return err
							}
						}
					}
				case object.BUILTIN_OBJ:
					for _, arg := range expr.Args {
						err := c.compile(arg, nil)
						if err != nil {
							return err
						}
					}
					c.emit(code.OpCall, len(expr.Args))

					variable := vars[n]
					err := c.assignValue(variable, nil)
					if err != nil {
						return err
					}
					n++
				}
			}
		case *ast.IndexExpr:
			symbol, err := c.compileIndexExpr(expr)
			if err != nil {
				return err
			}

			variable := vars[n]
			if symbol.Type.Type() == object.ARRAY_OBJ {
				defObj := object.GetDefaultObject(symbol.Type.(*object.Array).ElemType.String())
				err = c.assignValue(variable, &Symbol{Type: defObj})
				if err != nil {
					return err
				}
				n++
			} else if symbol.Type.Type() == object.HASH_OBJ {
				defObj := object.GetDefaultObject(symbol.Type.(*object.Hash).ValueType.String())
				err = c.assignValue(variable, &Symbol{Type: defObj})
				if err != nil {
					return err
				}

				if len(vars) == 1 {
					vars = append(vars, Variable{Name: "_", Type: VarIdent})
				}

				variable = vars[1]
				defObj = object.GetDefaultObject(object.BOOLEAN_OBJ.String())
				err = c.assignValue(variable, &Symbol{Type: defObj})
				if err != nil {
					return err
				}
				n += 2
			}
		case *ast.BinaryExpr:
			err := c.compileBinaryExpr(expr)
			if err != nil {
				return err
			}
			variable := vars[n]
			err = c.assignValue(variable, nil)
			if err != nil {
				return err
			}
			n++
		case *ast.FuncLit:
			fnSymbol, err := c.compileFuncLit(expr)
			if err != nil {
				return err
			}
			variable := vars[n]
			err = c.assignValue(variable, fnSymbol)
			if err != nil {
				return err
			}
			n++
		case *ast.CompositeLit:
			rtSymbol, err := c.compileCompositeLit(expr, nil)
			if err != nil {
				return err
			}
			variable := vars[n]
			err = c.assignValue(variable, rtSymbol)
			if err != nil {
				return err
			}
			n++
		default:
			line, column := parsePos(expr.Pos())
			return fmt.Errorf("%d:%d DefineStmt not support %T", line, column, expr)
		}
	}
	return nil
}

func (c *Compiler) compileBinaryAssign(node *ast.AssignStmt) error {
	node1 := ast.AssignStmt{Tok: token.ASSIGN}
	node1.Lhs = node.Lhs

	binNode := ast.BinaryExpr{}
	binNode.X = node.Lhs[0]
	binNode.Y = node.Rhs[0]
	binNode.Op = object.PairToken[node.Tok]

	node1.Rhs = []ast.Expr{&binNode}
	return c.compile(&node1, nil)
}

func (c *Compiler) compileIdent(node *ast.Ident) (Symbol, error) {
	if node.Name == "true" || node.Name == "false" {
		if node.Name == "true" {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
		return Symbol{Type: object.FALSE}, nil
	}

	symbol, ok := c.SymbolTable.Resolve(node.Name)
	if ok {
		c.loadSymbol(symbol)
		return symbol, nil
	} else {
		return symbol, fmt.Errorf("undefined: %s", node.Name)
	}
}

func (c *Compiler) compileBinaryExpr(node *ast.BinaryExpr) error {
	err := c.compile(node.X, nil)
	if err != nil {
		return err
	}

	err = c.compile(node.Y, nil)
	if err != nil {
		return err
	}

	switch node.Op {
	case token.ADD:
		c.emit(code.OpADD)
	case token.SUB:
		c.emit(code.OpSUB)
	case token.MUL:
		c.emit(code.OpMUL)
	case token.QUO:
		c.emit(code.OpQUO)
	case token.REM:
		c.emit(code.OpREM)
	case token.AND:
		c.emit(code.OpAND)
	case token.OR:
		c.emit(code.OpOR)
	case token.XOR:
		c.emit(code.OpXOR)
	case token.SHL:
		c.emit(code.OpSHL)
	case token.SHR:
		c.emit(code.OpSHR)
	case token.AND_NOT:
		c.emit(code.OpAND_NOT)
	case token.EQL:
		c.emit(code.OpEQL)
	case token.LSS:
		c.emit(code.OpLSS)
	case token.GTR:
		c.emit(code.OpGTR)
	case token.NEQ:
		c.emit(code.OpNEQ)
	case token.LEQ:
		c.emit(code.OpLEQ)
	case token.GEQ:
		c.emit(code.OpGEQ)
	case token.LAND:
		c.emit(code.OpLAND)
	case token.LOR:
		c.emit(code.OpLOR)
	default:
		return fmt.Errorf("binaryExpr not support %s", node.Op)
	}
	return nil
}

func (c *Compiler) compileCompositeLit(node *ast.CompositeLit, defaultObj object.Object) (*Symbol, error) {
	if node.Type != nil {
		switch ty := node.Type.(type) {
		case *ast.ArrayType:
			defObj := object.GetDefaultValueWithExpr(ty.Elt)
			for _, elt := range node.Elts {
				err := c.compile(elt, defObj)
				if err != nil {
					return nil, err
				}
			}
			c.emit(code.OpArray, len(node.Elts))
			symbol := Symbol{Type: &object.Array{ElemType: defObj.Type()}}
			return &symbol, nil
		case *ast.MapType:
			mm := &object.Hash{}

			defKObj := object.GetDefaultValueWithExpr(ty.Key)
			defVObj := object.GetDefaultValueWithExpr(ty.Value)
			if _, ok := defKObj.(object.Hashable); !ok {
				line, column := parsePos(ty.Key.Pos())
				return nil, fmt.Errorf("%d:%d key not a HashKey type", line, column)
			}
			mm.KeyType = defKObj.Type()
			mm.ValueType = defVObj.Type()

			for _, elt := range node.Elts {
				eltNode, _ := elt.(*ast.KeyValueExpr)
				if err := c.compile(eltNode.Key, defKObj); err != nil {
					return nil, err
				}
				if err := c.compile(eltNode.Value, defVObj); err != nil {
					return nil, err
				}
			}
			c.emit(code.OpHash, len(node.Elts)*2)
			return &Symbol{Type: mm}, nil
		}
	} else if node.Elts != nil {
		eltt := node.Elts[0]
		switch eltt.(type) {
		case *ast.KeyValueExpr:
			hashObj := defaultObj.(*object.Hash)
			for _, elt := range node.Elts {
				eltNode, _ := elt.(*ast.KeyValueExpr)
				if err := c.compile(eltNode.Key, object.GetDefaultObject(hashObj.KeyType.String())); err != nil {
					return nil, err
				}
				if err := c.compile(eltNode.Value, object.GetDefaultObject(hashObj.ValueType.String())); err != nil {
					return nil, err
				}
			}
			c.emit(code.OpHash, len(node.Elts)*2)
			return nil, nil
		default:
			if defaultObj.Type() == object.ARRAY_OBJ {
				for _, elt := range node.Elts {
					err := c.compile(elt, object.GetDefaultObject(defaultObj.(*object.Array).ElemType.String()))
					if err != nil {
						return nil, err
					}
				}
				c.emit(code.OpArray, len(node.Elts))
				return nil, nil
			}
		}
	}

	return nil, nil
}

func (c *Compiler) compileIndexExpr(node *ast.IndexExpr) (*Symbol, error) {
	symbol := Symbol{}
	switch x := node.X.(type) {
	case *ast.Ident:
		idtName := x.Name
		symbol, _ = c.SymbolTable.Resolve(idtName)
		c.loadSymbol(symbol)
	case *ast.IndexExpr:
		_, err := c.compileIndexExpr(node.X.(*ast.IndexExpr))
		if err != nil {
			return nil, err
		}
	case *ast.CompositeLit:
		_, err := c.compileCompositeLit(x, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("not support x node in IndexExpr")
	}

	err := c.compile(node.Index, nil)
	if err != nil {
		return nil, err
	}
	c.emit(code.OpIndex)
	return &symbol, nil
}

func (c *Compiler) compileFuncLit(node *ast.FuncLit) (*Symbol, error) {
	tmp := program.ParseFuncLit(node, nil)
	fn := tmp.(*object.Function)

	compiledFn, err := c.compileFunction(fn, "")
	if err != nil {
		return nil, err
	}

	fnIdx := c.addConstants(compiledFn)
	c.emit(code.OpClosure, fnIdx, compiledFn.FreeNum)
	fn.Body = nil
	fn.Env = nil
	return &Symbol{Type: fn}, nil
}

func (c *Compiler) compileFunction(fn *object.Function, fnName string) (*object.CompiledFunction, error) {
	c.enterScope()
	if fnName != "" {
		symbol, _ := c.SymbolTable.Resolve(fnName)
		c.SymbolTable.DefineFunctionName(fnName, symbol.Type)
	}

	numArgs, numResult := 0, 0
	for _, param := range fn.Params {
		c.SymbolTable.DefineWithType(param.Symbol.Name, object.GetDefaultValueFromElem(param.Type))
		numArgs++
	}
	for _, result := range fn.Results {
		if result.Symbol != nil {
			c.SymbolTable.DefineWithType(result.Symbol.Name, object.GetDefaultValueFromElem(result.Type))
			numResult++
		}
	}
	err := c.compileBlockStmt(fn.Body, nil)
	if err != nil {
		c.leaveScope()
		return nil, err
	}
	if !(c.lastInstructionIs(code.OpReturnValue) || c.lastInstructionIs(code.OpReturn)) {
		c.emit(code.OpReturn)
	}

	numLocals := c.SymbolTable.NumDefinitions
	freeSymbols := c.SymbolTable.FreeSymbols
	instructions := c.leaveScope()

	for _, s := range freeSymbols {
		c.loadSymbol(s)
	}

	compiledFn := &object.CompiledFunction{
		Instructions: instructions,
		NumLocals:    numLocals,
		NumParams:    numArgs,
		NumResult:    numResult,
		FreeNum:      len(freeSymbols),
	}

	return compiledFn, err
}

func (c *Compiler) compileIfStmt(node *ast.IfStmt) error {
	var existSymbol []Symbol
	var notExistSymbol []Symbol
	if node.Init != nil {
		init := node.Init.(*ast.AssignStmt)
		if init.Tok == token.DEFINE {
			for _, lh := range init.Lhs {
				ident, ok := lh.(*ast.Ident)
				if !ok {
					return fmt.Errorf("forStmt init is not *ast.Ident")
				}
				symbol, ok := c.SymbolTable.Resolve(ident.Name)
				if ok {
					existSymbol = append(existSymbol, symbol)
				} else {
					notExistSymbol = append(notExistSymbol, symbol)
				}
			}
		}
	}
	if len(existSymbol) > 0 {
		for _, symbol := range existSymbol {
			c.loadSymbol(symbol)
		}
		for _, symbol := range existSymbol {
			c.SymbolTable.DeleteSymbol(symbol.Name)
		}
	}

	if node.Init != nil {
		err := c.compile(node.Init, nil)
		if err != nil {
			return err
		}
	}

	err := c.compile(node.Cond, nil)
	if err != nil {
		return err
	}

	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 0)
	err = c.compile(node.Body, nil)
	if err != nil {
		return err
	}
	jumpPos := c.emit(code.OpJump, 0)
	afterConsequencePos := len(c.currentInstructions())
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if node.Else != nil {
		err = c.compile(node.Else, nil)
		if err != nil {
			return err
		}
	}
	afterConsequencePos = len(c.currentInstructions())
	c.changeOperand(jumpPos, afterConsequencePos)

	for _, symbol := range notExistSymbol {
		c.SymbolTable.DeleteSymbol(symbol.Name)
	}
	for _, symbol := range existSymbol {
		c.SymbolTable.DeleteSymbol(symbol.Name)
	}

	for _, symbol := range existSymbol {
		c.SymbolTable.ResetSymbol(symbol)
	}
	if len(existSymbol) > 0 {
		for i := len(existSymbol) - 1; i >= 0; i-- {
			c.storeSymbol(existSymbol[i])
		}
	}
	return nil
}

func (c *Compiler) compileForStmt(node *ast.ForStmt) error {
	var loop object.ForLoop

	c.enterScope()
	if node.Init != nil {
		init := node.Init.(*ast.AssignStmt)
		if init.Tok == token.DEFINE {
			for _, lh := range init.Lhs {
				ident, ok := lh.(*ast.Ident)
				if !ok {
					return fmt.Errorf("forStmt init is not *ast.Ident")
				}
				c.SymbolTable.Define(ident.Name)
			}
		}
		err := c.compile(node.Init, nil)
		if err != nil {
			return err
		}
	}
	loop.Init = c.currentInstructions()
	c.scopes[c.scopeIndex] = CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	err := c.compile(node.Cond, nil)
	if err != nil {
		return err
	}
	loop.Cond = c.currentInstructions()
	c.scopes[c.scopeIndex] = CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	err = c.compile(node.Body, nil)
	if err != nil {
		return err
	}
	loop.Body = c.currentInstructions()
	c.scopes[c.scopeIndex] = CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	err = c.compile(node.Post, nil)
	if err != nil {
		return err
	}
	loop.Post = c.currentInstructions()
	c.scopes[c.scopeIndex] = CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	numLocals := c.SymbolTable.NumDefinitions
	freeSymbols := c.SymbolTable.FreeSymbols
	c.leaveScope()
	for _, s := range freeSymbols {
		c.loadSymbol(s)
	}

	loop.NumLocals = numLocals
	loop.FreeNum = len(freeSymbols)

	fnIdx := c.addConstants(&loop)
	c.emit(code.OpClosure, fnIdx, loop.FreeNum)
	c.emit(code.OpForLoop)
	return nil
}

func (c *Compiler) compileRangeStmt(node *ast.RangeStmt) error {
	var rangeLoop object.RangeLoop

	c.enterScope()
	idtX, isIdent := node.X.(*ast.Ident)
	if !isIdent {
		idtX = ast.NewIdent("loop_x")
		var init ast.AssignStmt
		init.Lhs = []ast.Expr{idtX}
		init.Tok = token.DEFINE
		init.Rhs = []ast.Expr{node.X}
		err := c.compile(&init, nil)
		if err != nil {
			return err
		}
		rangeLoop.IsAnonymous = true
	}
	xSymbol, _ := c.SymbolTable.Resolve(idtX.Name)
	c.loadSymbol(xSymbol)
	rangeLoop.X = c.currentInstructions()
	c.scopes[c.scopeIndex] = CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.SymbolTable.Define("loop_K")
	c.SymbolTable.Define("loop_K")

	var body ast.BlockStmt
	if node.Key != nil {
		ident, _ := node.Key.(*ast.Ident)
		if ident.Name != "_" {
			body.List = append(body.List, &ast.AssignStmt{Lhs: []ast.Expr{node.Key}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.Ident{Name: "loop_K"}}})
		}
	}
	if node.Value != nil {
		ident, _ := node.Value.(*ast.Ident)
		if ident.Name != "_" {
			body.List = append(body.List, &ast.AssignStmt{Lhs: []ast.Expr{node.Key}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.Ident{Name: "loop_V"}}})
		}
	}
	body.List = append(body.List, node.Body.List...)
	err := c.compile(&body, nil)
	if err != nil {
		return err
	}
	rangeLoop.Body = c.currentInstructions()
	c.scopes[c.scopeIndex] = CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	numLocals := c.SymbolTable.NumDefinitions
	freeSymbols := c.SymbolTable.FreeSymbols
	c.leaveScope()
	for _, s := range freeSymbols {
		c.loadSymbol(s)
	}

	rangeLoop.NumLocals = numLocals
	rangeLoop.FreeNum = len(freeSymbols)

	fnIdx := c.addConstants(&rangeLoop)
	c.emit(code.OpClosure, fnIdx, rangeLoop.FreeNum)
	c.emit(code.OpRangeLoop)
	return nil
}

func (c *Compiler) compileBlockStmt(node *ast.BlockStmt, defaultType object.Object) error {
	for _, stmt := range node.List {
		err := c.compile(stmt, defaultType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileReturnStmt(node *ast.ReturnStmt) error {
	if node.Results == nil {
		c.emit(code.OpReturn)
		return nil
	}
	for _, result := range node.Results {
		err := c.compile(result, nil)
		if err != nil {
			return err
		}
	}
	n := len(node.Results)
	c.emit(code.OpReturnValue, n)
	return nil
}

func (c *Compiler) compileCallExpr(node *ast.CallExpr) error {
	fun := node.Fun
	switch fn := fun.(type) {
	case *ast.Ident:
		symbol, ok := c.SymbolTable.Resolve(fn.Name)
		if !ok {
			return fmt.Errorf("undefiend: %s", fn.Name)
		}
		c.loadSymbol(symbol)

		switch symbol.Type.Type() {
		case object.FUNCTION_OBJ:
			fnc := symbol.Type.(*object.Function)
			for _, arg := range node.Args {
				err := c.compile(arg, nil)
				if err != nil {
					return err
				}
			}
			nArgs := len(fnc.Params)
			nRet := 0
			for _, result := range fnc.Results {
				if result.Symbol != nil {
					obj := object.GetDefaultValueFromElem(result.Type)
					c.emit(code.OpConstant, c.addConstants(obj))
					nRet++
				}
			}
			c.emit(code.OpCall, nArgs+nRet)
		case object.BUILTIN_OBJ:
			for _, arg := range node.Args {
				err := c.compile(arg, nil)
				if err != nil {
					return err
				}
			}
			c.emit(code.OpCall, len(node.Args))
		}
	case *ast.FuncLit:
		fnObj := program.ParseFuncLit(fn, nil).(*object.Function)
		compiledFn, err := c.compileFunction(fnObj, "")
		if err != nil {
			return err
		}
		c.emit(code.OpClosure, c.addConstants(compiledFn), 0)

		nArgs := 0
		for _, arg := range node.Args {
			err = c.compile(arg, nil)
			if err != nil {
				return err
			}
			nArgs++
		}

		nRet := 0
		for _, result := range fnObj.Results {
			if result.Symbol != nil {
				obj := object.GetDefaultValueFromElem(result.Type)
				c.emit(code.OpConstant, c.addConstants(obj))
				nRet++
			}
		}

		c.emit(code.OpCall, nArgs+nRet)
	case *ast.CallExpr:
		err := c.compileCallExpr(fun.(*ast.CallExpr))
		if err != nil {
			return err
		}
		nArgs := 0
		for _, arg := range node.Args {
			err = c.compile(arg, nil)
			if err != nil {
				return err
			}
			nArgs++
		}
		c.emit(code.OpCall, nArgs)
	default:
		return fmt.Errorf("not support CallExpr: %T", fn)
	}
	return nil
}

func (c *Compiler) compileIncDecStmt(node *ast.IncDecStmt) error {
	err := c.compile(node.X, nil)
	if err != nil {
		return err
	}
	switch node.Tok {
	case token.INC:
		c.emit(code.OpINC)
	case token.DEC:
		c.emit(code.OpDEC)
	}
	ident, ok := node.X.(*ast.Ident)
	if ok {
		symbol, _ := c.SymbolTable.Resolve(ident.Name)
		c.storeSymbol(symbol)
	}
	return nil
}

func (c *Compiler) assignValue(v Variable, varSymbol *Symbol) error {
	switch v.Type {
	case VarIdent:
		symbol, ok := c.SymbolTable.Resolve(v.Name)
		if !ok {
			if varSymbol == nil {
				symbol = c.SymbolTable.Define(v.Name)
			} else {
				symbol = c.SymbolTable.DefineWithType(v.Name, varSymbol.Type)
			}
		}
		c.storeSymbol(symbol)
		return nil
	case VarIndex:
		symbol, _ := c.SymbolTable.Resolve(v.Name)
		c.loadSymbol(symbol)
		err := c.compile(v.Index, nil)
		if err != nil {
			return err
		}
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobalIndex, symbol.Index)
		} else if symbol.Scope == LocalScope {
			c.emit(code.OpSetLocalIndex, symbol.Index)
		}
		return nil
	default:
		return nil
	}
}

func parseBasicLit(node *ast.BasicLit) (object.Object, error) {
	switch node.Kind {
	case token.INT:
		value, err := strconv.ParseInt(node.Value, 10, 64)
		if err != nil {
			return nil, err
		} else {
			return &object.Int{Value: int(value)}, nil
		}
	case token.FLOAT:
		value, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return nil, err
		} else {
			return &object.Float64{Value: value}, nil
		}
	case token.STRING:
		value := parseString(node)
		return &object.String{Value: value}, nil
	case token.CHAR:
		value, err := parseChar(node)
		if err != nil {
			return nil, err
		} else {
			return &object.Byte{Value: value}, nil
		}
	default:
		return nil, fmt.Errorf("not support basic type %s", node.Kind.String())
	}
}

var escapeChMap = map[byte]byte{
	't':  '\t',
	'n':  '\n',
	'"':  '"',
	'\\': '\\',
	'r':  '\r',
}

func parseString(node *ast.BasicLit) string {
	str := node.Value[1 : len(node.Value)-1]
	var out bytes.Buffer
	for i := 0; i < len(str); i++ {
		cur := str[i]
		if cur == '\\' {
			next := str[i+1]
			if v, ok := escapeChMap[next]; ok {
				i += 1
				out.WriteByte(v)
			}
		} else if cur != '"' {
			out.WriteByte(cur)
		}
		if cur == '"' || cur == 0 {
			break
		}
	}
	return out.String()
}

func parseChar(node *ast.BasicLit) (uint8, error) {
	str := node.Value
	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") && len(str) == 3 {
		b := []byte(str)
		return b[1], nil
	}
	return 0, errors.New("")
}

func parsePos(p token.Pos) (int, int) {
	pos := tokenFile.Position(p)
	return pos.Line, pos.Column
}

type LhsItem struct {
	Name    string
	Scope   SymbolScope
	IsIndex bool
	Index   int64
	HashKey object.HashKey
}
