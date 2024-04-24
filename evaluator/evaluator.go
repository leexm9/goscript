package evaluator

import (
	"bytes"
	"errors"
	"go/ast"
	"go/token"
	"goscript/object"
	"goscript/program"
	"strconv"
	"strings"
)

func EvalProgram(prog *program.Program) object.Object {
	for _, stmt := range prog.Statements {
		obj := eval(stmt)
		return obj
	}
	return nil
}

func eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.ExprStmt:
		return eval(node.X)
	case *ast.BasicLit:
		return parseBasicLit(node)
	}
	return nil
}

func parseBasicLit(basic *ast.BasicLit) object.Object {
	switch basic.Kind {
	case token.INT:
		if value, err := strconv.ParseInt(basic.Value, 10, 64); err != nil {
			return object.NewError("%s cannot be represented by the type int", basic.Value)
		} else {
			return &object.Int{Value: int(value)}
		}
	case token.FLOAT:
		if value, err := strconv.ParseFloat(basic.Value, 64); err != nil {
			return object.NewError("%s cannot be represented by the type float", basic.Value)
		} else {
			return &object.Float64{Value: value}
		}
	case token.STRING:
		value := parseString(basic)
		return &object.String{Value: value}
	case token.CHAR:
		if value, err := parseChar(basic); err != nil {
			return object.NewError("%s cannot be represented by the type char", basic.Value)
		} else {
			return &object.Byte{Value: value}
		}
	default:
		return object.NewError("not support ast.BasicLit kind: %T", basic.Kind)
	}
}

// 转义符
var escapeChMap = map[byte]byte{
	't':  '\t',
	'n':  '\n',
	'"':  '"',
	'\\': '\\',
	'r':  '\r',
}

func parseString(basic *ast.BasicLit) string {
	str := basic.Value[1 : len(basic.Value)-1]
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

func parseChar(basic *ast.BasicLit) (uint8, error) {
	str := basic.Value
	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") && len(str) == 3 {
		b := []byte(str)
		return b[1], nil
	}
	return 0, errors.New("")
}
