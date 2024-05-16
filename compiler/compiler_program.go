package compiler

import (
	"go/ast"
	"goscript/code"
	"goscript/object"
	"goscript/program"
	"strings"
)

type VarType int

const (
	VarIdent VarType = iota
	VarIndex
	VarAttr
)

type Variable struct {
	Name      string
	Index     ast.Node
	Attribute ast.Node
	Type      VarType
}

func (c *Compiler) CompileProgram(prog *program.Program) error {
	tokenFile = prog.TokenFile
	store := prog.Env.GetStore()
	symbolTable := c.SymbolTable
	for name, value := range store {
		fn, ok := value.(*object.Function)
		if _, exist := symbolTable.Resolve(name); exist {
			continue
		}
		if ok {
		PRECOM:
			constantNum := len(c.constants)
			err := c.compileAstFunction(name, fn, symbolTable, constantNum)
			if err != nil {
				symbolTable.DeleteSymbol(name)
				c.constants = c.constants[:constantNum]

				msg := err.Error()
				if !strings.HasPrefix(msg, "undefined") {
					return err
				}
				idt := msg[11:]
				obj, ok := store[idt]
				if ok && obj.Type() == object.FUNCTION_OBJ {
					tmpFun, _ := obj.(*object.Function)
					err = c.compileAstFunction(idt, tmpFun, symbolTable, constantNum)
					if err != nil {
						return err
					} else {
						goto PRECOM
					}
				} else {
					return err
				}
			}
		}
	}
	c.globalDecls = prog.GlobalDecls

	num := len(prog.Statements)
	for i, stmt := range prog.Statements {
		err := c.compile(stmt, nil)
		if err != nil {
			return nil
		}
		switch stmt.(type) {
		case *ast.DeclStmt, *ast.AssignStmt:
		default:
			if num == i+1 {
				c.emit(code.OpPop)
			}
		}
	}
	return nil
}
