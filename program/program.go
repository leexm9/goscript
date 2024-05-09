package program

import (
	"fmt"
	"go/ast"
	"go/token"
	"goscript/object"
	"goscript/parser"
)

type Input struct {
	Name    string
	Content string
	IsStmt  bool
	IsCheck bool
}

type Program struct {
	Statements  []ast.Stmt
	Env         *object.Environment
	TokenFile   *token.File
	GlobalDecls int
}

func NewProgram() *Program {
	var prog Program
	env := object.NewEnvironment()
	prog.Env = env
	return &prog
}

var goTmpl = `package tmp
func main() {
%s
}`

func ParseFile(input Input) (*Program, error) {
	fset := token.NewFileSet()

	content := input.Content
	if input.IsStmt {
		content = fmt.Sprintf(goTmpl, content)
	}

	astFile, tokenFile, err := parser.ParseFile(fset, input.Name, content, 0)
	if err != nil {
		if input.IsStmt {
			err = formatError(err, 2)
		}
		return nil, err
	}

	if input.IsCheck {
		err = check(fset, astFile)
		if err != nil {
			if input.IsStmt {
				err = formatError(err, 2)
			}
			return nil, err
		}
	}

	prog := NewProgram()
	for _, decl := range astFile.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			name := decl.Name.Name
			if name == "main" {
				prog.Statements = decl.Body.List
			} else {
				addFunc(prog, name, decl)
				prog.GlobalDecls++
			}
		case *ast.GenDecl:
		default:
		}
	}

	prog.TokenFile = tokenFile
	return prog, nil
}

func addFunc(prog *Program, name string, funcDecl *ast.FuncDecl) {
	var funcLit ast.FuncLit
	funcLit.Type = funcDecl.Type
	funcLit.Body = funcDecl.Body
	function := ParseFuncLit(&funcLit, prog.Env)
	prog.Env.Set(name, function)
}

func ParseFuncLit(node *ast.FuncLit, env *object.Environment) object.Object {
	var fn object.Function

	if node.Type != nil {
		if node.Type.Params != nil {
			for _, field := range node.Type.Params.List {
				args := parseFunArgType(field)
				fn.Params = append(fn.Params, args...)
			}
		}
	}

	if node.Type.Results != nil {
		for _, field := range node.Type.Results.List {
			results := parseResultType(field)
			fn.Results = append(fn.Results, results...)
		}
	}

	fn.Body = node.Body
	fn.Env = env
	return &fn
}

func parseFunArgType(field *ast.Field) []object.FunArg {
	var args []object.FunArg

	var elemType object.ElemType
	switch ty := field.Type.(type) {
	case *ast.Ident:
		elemType = object.ElemType{Type: ty}
	case *ast.ArrayType:
		elemType = object.ElemType{Type: ty.Elt.(*ast.Ident), TypeElem: object.ElemArray}
	case *ast.MapType:
		elemType = object.ElemType{TypeElem: object.ElemHash}
		elemType.Types = make([]*ast.Ident, 2)
		elemType.Types[0] = ty.Key.(*ast.Ident)
		elemType.Types[1] = ty.Value.(*ast.Ident)
	default:
	}

	if field.Names != nil {
		for _, name := range field.Names {
			arg := object.FunArg{Symbol: name, Type: elemType}
			args = append(args, arg)
		}
	} else {
		args = append(args, object.FunArg{Type: elemType})
	}
	return args
}

func parseResultType(field *ast.Field) []object.FunResult {
	var results []object.FunResult

	ft := fmt.Sprintf("%T", field.Type)
	if ft == "*ast.FuncType" {
		var funResult object.FunResult
		funResult.IsFun = true
		funcType := field.Type.(*ast.FuncType)
		if funcType.Params != nil {
			for _, item := range funcType.Params.List {
				args := parseFunArgType(item)
				funResult.Params = append(funResult.Params, args...)
			}
		}
		if funcType.Results != nil {
			for _, item := range funcType.Results.List {
				rts := parseResultType(item)
				funResult.Results = append(funResult.Results, rts...)
			}
		}
		results = append(results, funResult)
	} else {
		var elemType object.ElemType
		switch ty := field.Type.(type) {
		case *ast.Ident:
			elemType = object.ElemType{Type: ty}
		case *ast.ArrayType:
			elemType = object.ElemType{Type: ty.Elt.(*ast.Ident), TypeElem: object.ElemArray}
		case *ast.MapType:
			elemType = object.ElemType{TypeElem: object.ElemHash}
			elemType.Types = make([]*ast.Ident, 2)
			elemType.Types[0] = ty.Key.(*ast.Ident)
			elemType.Types[1] = ty.Value.(*ast.Ident)
		default:
		}

		if field.Names != nil {
			for _, name := range field.Names {
				rt := object.FunResult{Symbol: name, Type: elemType}
				results = append(results, rt)
			}
		} else {
			rt := object.FunResult{Type: elemType}
			results = append(results, rt)
		}
	}
	return results
}
