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

var tokenFile *token.File

func EvalProgram(prog *program.Program) object.Object {
	tokenFile = prog.TokenFile
	var result object.Object
	for _, stmt := range prog.Statements {
		result = eval(stmt, prog.Env)

		switch rt := result.(type) {
		case *object.SingleReturn:
			return rt.Value
		case *object.MapExist:
			return rt.Value
		case *object.Error:
			return rt
		}
	}
	return result
}

func eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.ExprStmt:
		return eval(node.X, env)
	case *ast.ReturnStmt:
		return evalReturnStmt(node, env)
	case *ast.BlockStmt:
		return evalBlockStmt(node, env)
	case *ast.BinaryExpr:
		return evalBinaryExpr(node, env)
	case *ast.UnaryExpr:
		return evalUnaryExpr(node, env)
	case *ast.ParenExpr:
		return eval(node.X, env)
	case *ast.IndexExpr:
		return evalIndexExpr(node, env)
	case *ast.CallExpr:
		return evalCallExpr(node, env)
	case *ast.CompositeLit:
		return evalCompositeLit(node, env)
	case *ast.FuncLit:
		return evalFuncLit(node, env)
	case *ast.Ident:
		return evalIdentifier(node, env)
	case *ast.BasicLit:
		return parseBasicLit(node)
	case *ast.BranchStmt:
		if node.Tok == token.CONTINUE {
			return object.CONTINUE
		} else if node.Tok == token.BREAK {
			return object.BREAK
		} else {
			return object.NewError("evaluator: not support ast.BranchStmt %s", node.Tok)
		}
	default:
		return object.NewError("evaluator: not handle ast type %T", node)
	}
}

func evalReturnStmt(node *ast.ReturnStmt, env *object.Environment) object.Object {
	var objs []object.Object
	for _, result := range node.Results {
		obj := eval(result, env)
		if object.IsError(obj) {
			return obj
		}
		objs = append(objs, obj)
	}
	if node.Results == nil {
		return &object.SingleReturn{Value: nil}
	} else if len(node.Results) == 1 {
		switch rt := objs[0].(type) {
		case *object.SingleReturn:
			rt.FromFun = false
			return rt
		case *object.MultiReturn:
			rt.FromFun = false
			return rt
		default:
			return &object.SingleReturn{Value: objs[0]}
		}
	} else {
		return &object.MultiReturn{Values: objs}
	}
}

func evalBlockStmt(node *ast.BlockStmt, env *object.Environment) object.Object {
	for _, stmt := range node.List {
		obj := eval(stmt, env)
		if object.IsError(obj) {
			return obj
		}
		switch tmpRs := obj.(type) {
		case *object.SingleReturn:
			if !tmpRs.FromFun {
				return tmpRs
			}
		case *object.MultiReturn:
			if !tmpRs.FromFun {
				return tmpRs
			}
		case *object.Continue, *object.Break:
			return obj
		}
	}
	return nil
}

func evalCallExpr(node *ast.CallExpr, env *object.Environment) object.Object {
	args := evalExpressions(node.Args, env)
	if len(args) == 1 && object.IsError(args[0]) {
		return args[0]
	}

	line, column := parsePos(node.Pos())
	switch fnIdt := node.Fun.(type) {
	case *ast.Ident:
		fn := evalIdentifier(fnIdt, env)
		if object.IsError(fn) {
			return fn
		}

		switch function := fn.(type) {
		case *object.Function:
			extendEnv, err := extendFunctionEnv(function, args)
			if err != nil {
				return object.NewError("%d:%d %s to %s", line, column, err.Message, fnIdt.Name)
			}
			evaluated := eval(function.Body, extendEnv)
			return unwrapFuncReturn(evaluated, function)
		case *object.Builtin:
			if result := function.Fn(args...); result != nil {
				return result
			}
			return nil
		default:
			return object.NewError("%d:%d not a function %s", line, column, fn.Type())
		}
	case *ast.FuncLit:
		tmpFun := eval(fnIdt, env)
		function, ok := tmpFun.(*object.Function)
		if !ok {
			return object.NewError("%d:%d function literal error", line, column)
		}
		extendEnv, err := extendFunctionEnv(function, args)
		if err != nil {
			return object.NewError("%d:%d %s", line, column, err.Message)
		}
		evaluated := eval(function.Body, extendEnv)
		return unwrapFuncReturn(evaluated, function)
	default:
		return nil
	}
}

func evalExpressions(exprs []ast.Expr, env *object.Environment) []object.Object {
	var result []object.Object
	for _, expr := range exprs {
		evaluated := eval(expr, env)
		if object.IsError(evaluated) {
			return []object.Object{evaluated}
		}
		switch evaluated.Type() {
		case object.SINGLE_RETURN_OBJ:
			result = append(result, evaluated.(*object.SingleReturn).Value)
		case object.MULTI_RETURN_OBJ:
			multi := evaluated.(*object.MultiReturn)
			for _, value := range multi.Values {
				result = append(result, value)
			}
		case object.MAP_EXIST_OBJ:
			result = append(result, evaluated.(*object.MapExist).Value)
		default:
			result = append(result, evaluated)
		}
	}
	return result
}

func extendFunctionEnv(fn *object.Function, args []object.Object) (*object.Environment, *object.Error) {
	env := object.NewEnclosedEnvironment(fn.Env)
	if len(fn.Params) > len(args) {
		return env, object.NewError("not enough arguments in call")
	} else if len(fn.Params) < len(args) {
		return env, object.NewError("too many arguments in call")
	}
	for i, funArg := range fn.Params {
		defObj := object.GetDefaultValueFromElem(funArg.Type)
		if args[i].Type() != defObj.Type() {
			return env, object.NewError("cannot use '%s' (untyped %s constant) as %s value in argument", args[i], args[i].Type(), defObj.Type())
		}
		env.Set(funArg.Symbol.Name, args[i])
	}

	for _, funResult := range fn.Results {
		if funResult.Symbol != nil {
			defObj := object.GetDefaultValueFromElem(funResult.Type)
			env.Set(funResult.Symbol.Name, defObj)
		}
	}

	return env, nil
}

func unwrapFuncReturn(rt object.Object, fn *object.Function) object.Object {
	if object.IsError(rt) {
		return rt
	}
	if rtObj, ok := rt.(*object.SingleReturn); ok {
		if rtObj.Value != nil {
			mev, ok := rtObj.Value.(*object.MapExist)
			if ok {
				return &object.SingleReturn{Value: mev.Value, FromFun: true}
			} else {
				rtObj.FromFun = true
				return rtObj
			}
		} else if len(fn.Results) > 0 {
			var objs []object.Object
			for _, re := range fn.Results {
				obj, _ := fn.Env.Get(re.Symbol.Name)
				objs = append(objs, obj.Value)
			}
			if len(fn.Results) == 1 {
				return &object.SingleReturn{Value: objs[0], FromFun: true}
			} else {
				return &object.MultiReturn{Values: objs, FromFun: true}
			}
		}
	}
	return rt
}

func evalFuncLit(node *ast.FuncLit, env *object.Environment) object.Object {
	return program.ParseFuncLit(node, env)
}

func evalIndexExpr(node *ast.IndexExpr, env *object.Environment) object.Object {
	idt := eval(node.X, env)
	if object.IsError(idt) {
		return idt
	}
	idx := eval(node.Index, env)
	if object.IsError(idx) {
		return idx
	}
	return doIndex(idt, idx)
}

func doIndex(source object.Object, index object.Object) object.Object {
	switch source.Type() {
	case object.ARRAY_OBJ:
		array := source.(*object.Array)
		i := index.(object.Integer).Integer()
		return array.Elements[i]
	case object.HASH_OBJ:
		m := source.(*object.Hash)
		pair, ok := m.Pairs[index.(object.Hashable).HashKey()]
		if !ok {
			pair.Value = object.GetDefaultObject(m.ValueType.String())
		}
		return &object.MapExist{Value: pair.Value, Exist: ok}
	case object.STRING_OBJ:
		ss := source.(*object.String)
		i := index.(object.Integer).Integer()
		return &object.Uint8{Value: ss.Value[i]}
	case object.MAP_EXIST_OBJ:
		tmp := source.(*object.MapExist).Value
		return doIndex(tmp, index)
	default:
		return object.NewError("invalid operation: cannont index (variable of type %s)", source.Type())
	}
}

func evalCompositeLit(node *ast.CompositeLit, env *object.Environment) object.Object {
	if node.Type != nil {
		switch nodeType := node.Type.(type) {
		case *ast.ArrayType:
			var elems []object.Object
			defObj := object.GetDefaultValueWithExpr(nodeType.Elt)
			for _, elt := range node.Elts {
				obj := eval(elt, env)
				obj = object.ConvertValueWithType(obj, defObj)
				if object.IsError(obj) {
					line, column := parsePos(elt.Pos())
					return object.NewError("%d:%d cannot use (untyped '%s' constant) as %s value in array or slice literal", line, column, obj, defObj.Type())
				}
				elems = append(elems, obj)
			}
			var array object.Array
			if nodeType.Len == nil {
				array = object.NewArray(defObj.Type(), elems, false, -1)
			} else {
				ll := eval(nodeType.Len, env)
				if int(ll.(object.Integer).Integer()) != len(elems) {
					line, column := parsePos(node.Pos())
					return object.NewError("%d:%d out of bounds", line, column)
				}
				array = object.NewArray(defObj.Type(), elems, true, len(elems))
			}
			return &array
		case *ast.MapType:
			mm := &object.Hash{
				Pairs: make(map[object.HashKey]object.HashPair),
			}
			defKObj := object.GetDefaultValueWithExpr(nodeType.Key)
			defVObj := object.GetDefaultValueWithExpr(nodeType.Value)
			if _, ok := defKObj.(object.Hashable); !ok {
				line, column := parsePos(node.Pos())
				return object.NewError("%d:%d key not a Hashable type", line, column)
			}

			mm.KeyType = defKObj.Type()
			mm.ValueType = defVObj.Type()
			for _, elt := range node.Elts {
				line, column := parsePos(node.Pos())
				eltNode, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					return object.NewError("%d:%d not a ast.KeyValueExpr", line, column)
				}
				keyVal := eval(eltNode.Key, env)
				keyVal = object.ConvertValueWithType(keyVal, defKObj)
				if object.IsError(keyVal) {
					return keyVal
				}
				valVal := eval(eltNode.Value, env)
				valVal = object.ConvertValueWithType(valVal, defVObj)
				if object.IsError(valVal) {
					return valVal
				}

				pair := object.HashPair{Key: keyVal, Value: valVal}
				mm.Pairs[keyVal.(object.Hashable).HashKey()] = pair
			}
			return mm
		}
	} else if node.Elts != nil {
		eltt := node.Elts[0]
		switch eltt.(type) {
		case *ast.KeyValueExpr:
			hash := &object.Hash{}
			hash.Pairs = make(map[object.HashKey]object.HashPair)
			for _, elt := range node.Elts {
				eltKV := elt.(*ast.KeyValueExpr)
				k := eval(eltKV.Key, env)
				v := eval(eltKV.Value, env)
				hash.Pairs[k.(object.Hashable).HashKey()] = object.HashPair{Key: k, Value: v}
			}
			return hash
		default:
			var array []object.Object
			for _, elt := range node.Elts {
				obj := eval(elt, env)
				array = append(array, obj)
			}
			return &object.Array{Elements: array}
		}
	}
	return nil
}

func evalIdentifier(node *ast.Ident, env *object.Environment) object.Object {
	if node.Name == "true" {
		return object.TRUE
	} else if node.Name == "false" {
		return object.FALSE
	} else {
		val, ok := env.Get(node.Name)
		if ok {
			return val.GetValue()
		}
		builtin := object.GetBuiltinByName(node.Name)
		if builtin != nil {
			return builtin
		}
	}
	line, column := parsePos(node.Pos())
	return object.NewError("%d:%d ident not found: %s", line, column, node.Name)
}

func evalUnaryExpr(node *ast.UnaryExpr, env *object.Environment) object.Object {
	line, column := parsePos(node.X.Pos())
	obj := eval(node.X, env)
	if object.IsError(obj) {
		return obj
	}
	switch node.Op {
	case token.NOT:
		if obj == object.TRUE {
			return object.FALSE
		} else if obj == object.FALSE {
			return object.TRUE
		} else {
			return object.NewError("%d:%d operator ! not defined on %s", line, column, obj.Type())
		}
	case token.SUB:
		if obj.Type().IsInteger() {
			return object.ConvertToInt(obj.Type(), -obj.(object.Integer).Integer())
		} else if obj.Type().IsFloat() {
			return object.ConvertToFloat(obj.Type(), -obj.(object.Float).Float())
		} else {
			return object.NewError("%d:%d operator - not defined on %s", line, column, obj.Type())
		}
	case token.ADD:
		if obj.Type().IsInteger() || obj.Type().IsFloat() {
			return obj
		} else {
			return object.NewError("%d:%d operator + not defined on %s", line, column, obj.Type())
		}
	default:
		return object.NewError("%d:%d operator %s not support on %s", line, column, node.Op, obj.Type())
	}
}

func evalBinaryExpr(node *ast.BinaryExpr, env *object.Environment) object.Object {
	leftObj := eval(node.X, env)
	if object.IsError(leftObj) {
		return leftObj
	}
	rightObj := eval(node.Y, env)
	if object.IsError(rightObj) {
		return rightObj
	}
	obj := handleBinaryExpr(node.Op, leftObj, rightObj)
	if object.IsError(obj) {
		line, column := parsePos(node.Pos())
		return object.NewError("%d:%d %s", line, column, obj)
	}
	return obj
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

func handleBinaryExpr(op token.Token, left, right object.Object) object.Object {
	switch left.Type() {
	case object.SINGLE_RETURN_OBJ:
		left = left.(*object.SingleReturn).Value
	case object.MAP_EXIST_OBJ:
		left = left.(*object.MapExist).Value
	default:
	}

	switch right.Type() {
	case object.SINGLE_RETURN_OBJ:
		right = right.(*object.SingleReturn).Value
	case object.MAP_EXIST_OBJ:
		right = right.(*object.MapExist).Value
	default:
	}

	if left.Type() != right.Type() {
		return object.NewError("mismatched types %s and %s", left.Type(), right.Type())
	}
	switch {
	case left.Type().IsInteger():
		return handleIntegerBinaryExpr(op, left, right)
	case left.Type().IsFloat():
		return handleFloatBinaryExpr(op, left, right)
	case left.Type() == object.BOOLEAN_OBJ:
		return handleBooleanBinaryExpr(op, left, right)
	case left.Type() == object.STRING_OBJ:
		return handleStringBinaryExpr(op, left, right)
	default:
		return object.NewError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func handleIntegerBinaryExpr(op token.Token, left, right object.Object) object.Object {
	valType := left.Type()
	leftVal := left.(object.Integer).Integer()
	rightVal := right.(object.Integer).Integer()
	switch op {
	case token.ADD:
		return object.ConvertToInt(valType, leftVal+rightVal)
	case token.SUB:
		return object.ConvertToInt(valType, leftVal-rightVal)
	case token.MUL:
		return object.ConvertToInt(valType, leftVal*rightVal)
	case token.QUO:
		return object.ConvertToInt(valType, leftVal/rightVal)
	case token.REM:
		return object.ConvertToInt(valType, leftVal%rightVal)
	case token.AND:
		return object.ConvertToInt(valType, leftVal&rightVal)
	case token.OR:
		return object.ConvertToInt(valType, leftVal|rightVal)
	case token.XOR:
		return object.ConvertToInt(valType, leftVal^rightVal)
	case token.SHL:
		return object.ConvertToInt(valType, leftVal<<rightVal)
	case token.SHR:
		return object.ConvertToInt(valType, leftVal>>rightVal)
	case token.AND_NOT:
		return object.ConvertToInt(valType, leftVal&^rightVal)
	case token.EQL:
		return object.ConvertToBoolean(leftVal == rightVal)
	case token.LSS:
		return object.ConvertToBoolean(leftVal < rightVal)
	case token.GTR:
		return object.ConvertToBoolean(leftVal > rightVal)
	case token.NEQ:
		return object.ConvertToBoolean(leftVal != rightVal)
	case token.LEQ:
		return object.ConvertToBoolean(leftVal <= rightVal)
	case token.GEQ:
		return object.ConvertToBoolean(leftVal >= rightVal)
	default:
		return object.NewError("the operator %s is not defined on %s", op, valType)
	}
}

func handleFloatBinaryExpr(op token.Token, left, right object.Object) object.Object {
	valType := left.Type()
	leftVal := left.(object.Float).Float()
	rightVal := right.(object.Float).Float()
	switch op {
	case token.ADD:
		return object.ConvertToFloat(valType, leftVal+rightVal)
	case token.SUB:
		return object.ConvertToFloat(valType, leftVal-rightVal)
	case token.MUL:
		return object.ConvertToFloat(valType, leftVal*rightVal)
	case token.QUO:
		return object.ConvertToFloat(valType, leftVal/rightVal)
	case token.EQL:
		return object.ConvertToBoolean(leftVal == rightVal)
	case token.LSS:
		return object.ConvertToBoolean(leftVal < rightVal)
	case token.GTR:
		return object.ConvertToBoolean(leftVal > rightVal)
	case token.NEQ:
		return object.ConvertToBoolean(leftVal != rightVal)
	case token.LEQ:
		return object.ConvertToBoolean(leftVal <= rightVal)
	case token.GEQ:
		return object.ConvertToBoolean(leftVal >= rightVal)
	default:
		return object.NewError("the operator %s is not defined on %s", op, valType)
	}
}

func handleBooleanBinaryExpr(op token.Token, left, right object.Object) object.Object {
	leftVal := left.(*object.Boolean).Value
	rightVal := right.(*object.Boolean).Value
	switch op {
	case token.LAND:
		return object.ConvertToBoolean(leftVal && rightVal)
	case token.LOR:
		return object.ConvertToBoolean(leftVal || rightVal)
	case token.EQL:
		return object.ConvertToBoolean(leftVal == rightVal)
	case token.NEQ:
		return object.ConvertToBoolean(leftVal != rightVal)
	default:
		return object.NewError("the operator %s is not defined on %s", op, left.Type())
	}
}

func handleStringBinaryExpr(op token.Token, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	if op == token.ADD {
		return &object.String{Value: leftVal + rightVal}
	} else {
		cp := strings.Compare(leftVal, rightVal)
		switch op {
		case token.EQL:
			if cp == 0 {
				return object.TRUE
			} else {
				return object.FALSE
			}
		case token.NEQ:
			if cp != 0 {
				return object.TRUE
			} else {
				return object.FALSE
			}
		default:
			return object.NewError("the operator %s is not defined on %s", op, left.Type())
		}
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

func parsePos(p token.Pos) (int, int) {
	pos := tokenFile.Position(p)
	return pos.Line, pos.Column
}
