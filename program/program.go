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
	prog.Env.Set(name, funcLit)
}
