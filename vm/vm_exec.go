package vm

import (
	"errors"
	"fmt"
	"goscript/code"
	"goscript/object"
	"strings"
)

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().Ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().Ip++

		ip = vm.currentFrame().Ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpPop:
			vm.execPop()
		case code.OpTrue:
			err := vm.push(object.TRUE)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(object.FALSE)
			if err != nil {
				return err
			}
		case code.OpConstant:
			idx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().Ip += 2
			err := vm.push(vm.constants[idx])
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIdx := code.ReadUint16(ins[ip+1:])
			numFrees := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().Ip += 3
			err := vm.pushClosure(int(constIdx), int(numFrees))
			if err != nil {
				return err
			}
		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().Cl
			err := vm.push(currentClosure)
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().Ip += 1
			definition := object.Builtins[builtinIdx]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpADD, code.OpSUB, code.OpMUL, code.OpQUO, code.OpREM,
			code.OpAND, code.OpOR, code.OpXOR, code.OpSHL, code.OpSHR, code.OpAND_NOT,
			code.OpEQL, code.OpLSS, code.OpGTR,
			code.OpNEQ, code.OpLEQ, code.OpGEQ,
			code.OpLAND, code.OpLOR:
			right := vm.pop()
			left := vm.pop()

			rt := doBinaryExpr(op, left, right)
			if object.IsError(rt) {
				return errors.New(rt.(*object.Error).Message)
			}
			err := vm.push(rt)
			if err != nil {
				return err
			}
		case code.OpPrefixSub, code.OpPrefixAdd, code.OpNOT, code.OpINC, code.OpDEC:
			obj := vm.pop()
			rt := doUnaryExpr(op, obj)
			if object.IsError(rt) {
				return errors.New(rt.(*object.Error).Message)
			}
			err := vm.push(rt)
			if err != nil {
				return err
			}
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().Ip += 2

			condition := vm.pop()
			if !object.IsTruthy(condition) {
				vm.currentFrame().Ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().Ip = pos - 1
		case code.OpGetGlobal:
			idx := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().Ip += 2
			err := vm.push(vm.globals[idx])
			if err != nil {
				return err
			}
		case code.OpGetLocal:
			localIdx := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().Ip += 2
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.BasePointer+int(localIdx)])
			if err != nil {
				return err
			}
		case code.OpGetFree:
			freeIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().Ip += 1

			currentClosure := vm.currentFrame().currentClosure()
			err := vm.push(currentClosure.Free[freeIdx])
			if err != nil {
				return err
			}
		case code.OpSetGlobal, code.OpSetLocal:
			tip, err := vm.execSetGlobalLocal(op, ins, ip)
			if err != nil {
				return err
			}
			vm.currentFrame().Ip = tip
		case code.OpSetFree:
			tip, err := vm.execSetFree(ins, ip)
			if err != nil {
				return err
			}
			vm.currentFrame().Ip = tip
		case code.OpSetNil:
			err := vm.execSetNil()
			if err != nil {
				return err
			}
		case code.OpNull:
			err := vm.push(object.NULL)
			if err != nil {
				return err
			}
		case code.OpArray:
			nums := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().Ip += 2

			array := vm.buildArray(vm.sp-nums, vm.sp)
			vm.sp = vm.sp - nums

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			nums := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().Ip += 2

			hash, err := vm.buildHash(vm.sp-nums, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - nums

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.execIndexExpr(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().Ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpForLoop:
			err := vm.execForLoop()
			if err != nil {
				return err
			}
		case code.OpRangeLoop:
			err := vm.execRangeLoop()
			if err != nil {
				return err
			}
		case code.OpSetGlobalIndex, code.OpSetLocalIndex:
			tip, err := vm.execSetIndex(op, ins, ip)
			if err != nil {
				return err
			}
			vm.currentFrame().Ip = tip
		case code.OpReturnValue:
			err := vm.execReturnValue(ins, ip)
			if err != nil {
				return err
			}
		case code.OpReturn:
			err := vm.execReturn()
			if err != nil {
				return err
			}
		case code.OpContinue, code.OpBreak:
			tip, err := vm.execContinueOrBreak(op)
			if err != nil {
				return err
			}
			vm.currentFrame().Ip = tip
		}
	}
	return nil
}

func (vm *VM) pushClosure(idx, numFrees int) error {
	constant := vm.constants[idx]

	free := make([]object.Object, numFrees)
	for i := 0; i < numFrees; i++ {
		free[i] = vm.stack[vm.sp-numFrees+i]
	}
	vm.sp = vm.sp - numFrees

	var closure *object.Closure
	switch constant.(type) {
	case *object.CompiledFunction:
		fn, _ := constant.(*object.CompiledFunction)
		closure = &object.Closure{Fn: fn, Free: free}
	case *object.ForLoop:
		loop, _ := constant.(*object.ForLoop)
		closure = &object.Closure{ForLoop: loop, Free: free}
	case *object.RangeLoop:
		rangeLoop, _ := constant.(*object.RangeLoop)
		closure = &object.Closure{ForLoop: rangeLoop, Free: free}
	default:
		return fmt.Errorf("not a function: %+v", constant)
	}
	return vm.push(closure)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	case *object.SingleReturn:
		tmp := callee.Value.(*object.Closure)
		vm.stack[vm.sp-1-numArgs] = tmp
		return vm.executeCall(numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1
	if result != nil {
		return vm.push(result)
	}
	return nil
}

func (vm *VM) buildArray(startIdx, endIdx int) object.Object {
	elements := make([]object.Object, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		elements[i-startIdx] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIdx, endIdx int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)
	for i := startIdx; i < endIdx; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := pair.Key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		pairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: pairs}, nil
}

func (vm *VM) execIndexExpr(left object.Object, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INT_OBJ:
		return vm.execArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.execHashIndex(left, index)
	case left.Type() == object.MAP_EXIST_OBJ:
		return vm.execIndexExpr(left.(*object.MapExist).Value, index)
	case left.Type() == object.STRING_OBJ:
		return vm.execStringIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) execArrayIndex(left, index object.Object) error {
	arrayObj := left.(*object.Array)
	idx := index.(*object.Int).Value

	maxIdx := len(arrayObj.Elements) - 1
	if idx < 0 {
		return fmt.Errorf("index out of range [%d]", idx)
	}
	if idx > maxIdx {
		return fmt.Errorf("index out of range [%d] with length %d", idx, maxIdx)
	}

	return vm.push(arrayObj.Elements[idx])
}

func (vm *VM) execHashIndex(left, index object.Object) error {
	hashObj := left.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	var mapExist object.MapExist
	mapExist.Exist = ok
	if ok {
		mapExist.Value = pair.Value
	} else {
		mapExist.Value = object.NULL
	}
	return vm.push(&mapExist)
}

func (vm *VM) execStringIndex(left, index object.Object) error {
	str := left.(*object.String).Value
	idx := index.(*object.Int).Value

	maxIdx := len(str) - 1
	if idx < 0 {
		return fmt.Errorf("index out of range [%d]", idx)
	}
	if idx > maxIdx {
		return fmt.Errorf("index out of range [%d] with length %d", idx, maxIdx)
	}
	return vm.push(&object.Uint8{Value: str[idx]})
}

func (vm *VM) execReturnValue(ins code.Instructions, ip int) error {
	frame := vm.currentFrame()

	num := int(code.ReadUint8(ins[ip+1:]))
	rts := make([]object.Object, num)
	for i := num - 1; i >= 0; i-- {
		rts[i] = vm.pop()
	}

	var rt object.Object
	if len(rts) > 1 {
		rt = &object.MultiReturn{Values: rts, FromFun: false}
	} else {
		switch tmp := rts[0].(type) {
		case *object.SingleReturn, *object.MultiReturn:
			rt = tmp
		default:
			rt = &object.SingleReturn{Value: tmp, FromFun: true}
		}
	}

	// 特殊处理
	if !frame.IsMain {
		vm.stack[frame.BasePointer-1] = rt
		vm.sp = frame.BasePointer
		vm.popFrame()
	} else {
		vm.stack[0] = rt
		vm.sp = 1
		frame.Ip = len(frame.Instructions()) - 2
	}
	return nil
}

func (vm *VM) execReturn() error {
	frame := vm.popFrame()

	numArgs := frame.Cl.Fn.NumParams
	numResult := frame.Cl.Fn.NumResult
	rts := make([]object.Object, numResult)
	if numResult > 0 {
		for i := 0; i < numResult; i++ {
			rts[i] = vm.stack[frame.BasePointer+numArgs+i]
		}
	}
	if len(rts) == 0 {
		vm.sp = frame.BasePointer - 1
	} else {
		var rt object.Object
		if len(rts) > 1 {
			rt = &object.MultiReturn{Values: rts, FromFun: true}
		} else {
			switch tmp := rts[0].(type) {
			case *object.SingleReturn, *object.MultiReturn:
				rt = tmp
			default:
				rt = &object.SingleReturn{Value: tmp, FromFun: true}
			}
		}
		vm.stack[frame.BasePointer-1] = rt
		vm.sp = frame.BasePointer
	}
	return nil
}

func (vm *VM) execPop() {
	obj := vm.pop()
	switch rt := obj.(type) {
	case *object.SingleReturn:
		vm.stack[vm.sp] = rt.Value
	case *object.MapExist:
		vm.stack[vm.sp] = rt.Value
	default:
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParams+cl.Fn.NumResult {
		return fmt.Errorf("execute function wrong number of arguments: want=%d, got=%d", cl.Fn.NumParams, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.BasePointer + cl.Fn.NumLocals
	return nil
}

func (vm *VM) execForLoop() error {
	pos := vm.sp - 1

	closure := vm.stack[pos].(*object.Closure)
	forLoop := closure.ForLoop.(*object.ForLoop)

	var initFrame, condFrame, bodyFrame, postFrame *Frame
	if forLoop.Init != nil {
		initFn := &object.CompiledFunction{
			Instructions: forLoop.Init,
			NumLocals:    forLoop.NumLocals,
			NumParams:    0,
			FreeNum:      forLoop.FreeNum,
		}
		initFrame = NewFrame(&object.Closure{Fn: initFn, Free: closure.Free}, vm.sp)
	}

	condFn := &object.CompiledFunction{
		Instructions: forLoop.Cond,
		NumLocals:    forLoop.NumLocals,
		NumParams:    0,
		FreeNum:      forLoop.FreeNum,
	}
	condFrame = NewFrame(&object.Closure{Fn: condFn, Free: closure.Free}, vm.sp)

	bodyFn := &object.CompiledFunction{
		Instructions: forLoop.Body,
		NumLocals:    forLoop.NumLocals,
		NumParams:    0,
		FreeNum:      forLoop.FreeNum,
	}
	bodyFrame = NewFrame(&object.Closure{Fn: bodyFn, Free: closure.Free}, vm.sp)

	postFn := &object.CompiledFunction{
		Instructions: forLoop.Post,
		NumLocals:    forLoop.NumLocals,
		NumParams:    0,
		FreeNum:      forLoop.FreeNum,
	}
	postFrame = NewFrame(&object.Closure{Fn: postFn, Free: closure.Free}, vm.sp)

	if initFrame != nil {
		vm.pushFrame(initFrame)
		vm.sp = initFrame.BasePointer + forLoop.NumLocals
		err := vm.Run()
		if err != nil {
			return err
		}
		vm.popFrame()
	}

	vm.pushFrame(condFrame)
	vm.sp = condFrame.BasePointer + forLoop.NumLocals
	err := vm.Run()
	if err != nil {
		return err
	}
	vm.popFrame()

	cond := vm.pop()
	for object.IsTruthy(cond) {
		bodyFrame.Ip = -1
		vm.pushFrame(bodyFrame)
		err = vm.Run()
		if err != nil {
			return err
		}
		vm.popFrame()

		tmpObj := vm.stack[vm.sp-1]
		objType := tmpObj.Type()
		if objType > 15 && objType < 21 {
			vm.pop()
			if objType == object.BREAK_OBJ {
				break
			} else if objType < 19 {
				curFrame := vm.popFrame()
				vm.stack[curFrame.BasePointer-1] = tmpObj
				vm.sp = curFrame.BasePointer
				return nil
			}
		}

		postFrame.Ip = -1
		vm.pushFrame(postFrame)
		err = vm.Run()
		if err != nil {
			return err
		}
		vm.popFrame()

		condFrame.Ip = -1
		vm.pushFrame(condFrame)
		err = vm.Run()
		if err != nil {
			return err
		}
		vm.popFrame()

		cond = vm.pop()
	}

	vm.sp = pos
	if bodyFn.FreeNum > 0 {
		for i := 0; i < bodyFn.FreeNum; i++ {
			vm.stack[pos-1-i] = bodyFrame.Cl.Free[bodyFn.FreeNum-1-i]
		}
	}
	return nil
}

func (vm *VM) execRangeLoop() error {
	pos := vm.sp - 1

	closure := vm.stack[pos].(*object.Closure)
	rangeLoop := closure.ForLoop.(*object.RangeLoop)

	var xFrame, bodyFrame *Frame
	xFn := &object.CompiledFunction{
		Instructions: rangeLoop.X,
		NumLocals:    rangeLoop.NumLocals,
		NumParams:    0,
		FreeNum:      rangeLoop.FreeNum,
	}
	xFrame = NewFrame(&object.Closure{Fn: xFn, Free: closure.Free}, vm.sp)

	bodyFn := &object.CompiledFunction{
		Instructions: rangeLoop.Body,
		NumLocals:    rangeLoop.NumLocals,
		NumParams:    0,
		FreeNum:      rangeLoop.FreeNum,
	}
	bodyFrame = NewFrame(&object.Closure{Fn: bodyFn, Free: closure.Free}, vm.sp)

	vm.pushFrame(xFrame)
	err := vm.Run()
	if err != nil {
		return err
	}
	vm.popFrame()

	obj := vm.pop()
	switch obj := obj.(type) {
	case *object.Array:
		for i, elem := range obj.Elements {
			if rangeLoop.IsAnonymous {
				vm.stack[bodyFrame.BasePointer+1] = &object.Int{Value: i}
				vm.stack[bodyFrame.BasePointer+2] = elem
			} else {
				vm.stack[bodyFrame.BasePointer+0] = &object.Int{Value: i}
				vm.stack[bodyFrame.BasePointer+1] = elem
			}

			bodyFrame.Ip = -1
			vm.pushFrame(bodyFrame)
			vm.sp = bodyFrame.BasePointer + rangeLoop.NumLocals
			err := vm.Run()
			if err != nil {
				return err
			}
			vm.popFrame()

			tmpObj := vm.stack[vm.sp-1]
			objType := tmpObj.Type()
			if objType > 15 && objType < 21 {
				vm.pop()
				if objType == object.BREAK_OBJ {
					break
				} else if objType < 19 {
					curFrame := vm.popFrame()
					vm.stack[curFrame.BasePointer-1] = tmpObj
					vm.sp = curFrame.BasePointer
					return nil
				}
			}
		}
	case *object.Hash:
		for _, pair := range obj.Pairs {
			if rangeLoop.IsAnonymous {
				vm.stack[bodyFrame.BasePointer+1] = pair.Key
				vm.stack[bodyFrame.BasePointer+2] = pair.Value
			} else {
				vm.stack[bodyFrame.BasePointer+0] = pair.Key
				vm.stack[bodyFrame.BasePointer+1] = pair.Value
			}

			bodyFrame.Ip = -1
			vm.pushFrame(bodyFrame)
			vm.sp = bodyFrame.BasePointer + rangeLoop.NumLocals
			err := vm.Run()
			if err != nil {
				return err
			}
			vm.popFrame()

			tmpObj := vm.stack[vm.sp-1]
			objType := tmpObj.Type()
			if objType > 15 && objType < 21 {
				vm.pop()
				if objType == object.BREAK_OBJ {
					break
				} else if objType < 19 {
					curFrame := vm.popFrame()
					vm.stack[curFrame.BasePointer-1] = tmpObj
					vm.sp = curFrame.BasePointer
					return nil
				}
			}
		}
	}

	vm.sp = pos
	if bodyFn.FreeNum > 0 {
		for i := 0; i < bodyFn.FreeNum; i++ {
			vm.stack[pos-1-i] = bodyFrame.Cl.Free[bodyFn.FreeNum-1-i]
		}
	}
	return nil
}

func (vm *VM) execContinueOrBreak(op code.Opcode) (int, error) {
	frame := vm.currentFrame()
	if frame.IsLoop {
		if op == code.OpContinue {
			vm.stack[frame.BasePointer+frame.Cl.Fn.NumLocals] = object.CONTINUE
		} else {
			vm.stack[frame.BasePointer+frame.Cl.Fn.NumLocals] = object.BREAK
		}
	}
	vm.sp = frame.BasePointer + frame.Cl.Fn.NumLocals + 1
	return len(frame.Cl.Fn.Instructions) - 1, nil
}

func (vm *VM) execSetGlobalLocal(op code.Opcode, ins code.Instructions, ip int) (int, error) {
	idx := int(code.ReadUint16(ins[ip+1:]))
	ip += 2

	pos := vm.sp - 1
	obj := vm.stack[pos]
	newValue, obj, needPop := extractData(obj)
	if needPop {
		vm.pop()
	} else {
		vm.stack[pos] = obj
	}

	if op == code.OpSetGlobal {
		vm.globals[idx] = newValue
	} else {
		frame := vm.currentFrame()
		vm.stack[frame.BasePointer+idx] = newValue
	}
	return ip, nil
}

func (vm *VM) execSetNil() error {
	pos := vm.sp - 1
	obj := vm.stack[pos]
	_, obj, needPop := extractData(obj)
	if needPop {
		vm.pop()
	} else {
		vm.stack[pos] = obj
	}
	return nil
}

func (vm *VM) execSetFree(ins code.Instructions, ip int) (int, error) {
	idx := int(code.ReadUint8(ins[ip+1:]))
	ip += 1

	frame := vm.currentFrame()
	free := frame.currentClosure().Free

	pos := vm.sp - 1
	obj := vm.stack[pos]
	newValue, obj, needPop := extractData(obj)
	if needPop {
		vm.pop()
	} else {
		vm.stack[pos] = obj
	}

	free[idx] = newValue
	frame.Cl.Free = free
	vm.stack[frame.BasePointer-1] = frame.Cl
	return ip, nil
}

func (vm *VM) execSetIndex(op code.Opcode, ins code.Instructions, ip int) (int, error) {
	idx := int(code.ReadUint16(ins[ip+1:]))
	ip += 2

	frame := vm.currentFrame()
	idxObj := vm.pop()
	complexObj := vm.pop()

	pos := vm.sp - 1
	newObj := vm.stack[pos]
	newValue, newObj, needPop := extractData(newObj)
	if needPop {
		vm.pop()
	} else {
		vm.stack[pos] = newObj
	}

	switch cobj := complexObj.(type) {
	case *object.Array:
		index := idxObj.(*object.Int).Value
		cobj.Elements[index] = newValue
		if op == code.OpSetGlobalIndex {
			vm.globals[idx] = cobj
		} else {
			vm.stack[frame.BasePointer+idx] = cobj
		}
	case *object.Hash:
		index := idxObj.(object.Hashable).HashKey()
		_, ok := cobj.Pairs[index]
		if ok {
			cobj.Pairs[index] = object.HashPair{Key: idxObj, Value: newValue}
		} else {
			pair := object.HashPair{Key: idxObj, Value: newValue}
			cobj.Pairs[index] = pair
		}
		if op == code.OpSetGlobalIndex {
			vm.globals[idx] = cobj
		} else {
			vm.stack[frame.BasePointer+idx] = cobj
		}
	}
	return ip, nil
}

func doUnaryExpr(op code.Opcode, obj object.Object) object.Object {
	switch op {
	case code.OpNOT:
		if obj == object.TRUE {
			return object.FALSE
		} else if obj == object.FALSE {
			return object.TRUE
		} else {
			return object.NewError("operator ! not defined on %s", obj.Type())
		}
	case code.OpPrefixSub:
		if obj.Type().IsInteger() {
			return object.ConvertToInt(obj.Type(), -obj.(object.Integer).Integer())
		} else if obj.Type().IsFloat() {
			return object.ConvertToFloat(obj.Type(), -obj.(object.Float).Float())
		} else {
			return object.NewError("operator - not defined on %s", obj.Type())
		}
	case code.OpPrefixAdd:
		if obj.Type().IsInteger() || obj.Type().IsFloat() {
			return obj
		} else {
			return object.NewError("operator + not defined on %s", obj.Type())
		}
	case code.OpINC:
		if obj.Type().IsInteger() {
			return object.ConvertToInt(obj.Type(), obj.(object.Integer).Integer()+int64(1))
		} else if obj.Type().IsFloat() {
			return object.ConvertToFloat(obj.Type(), -obj.(object.Float).Float()+float64(1))
		} else {
			return object.NewError("operator ++ not defined on %s", obj.Type())
		}
	case code.OpDEC:
		if obj.Type().IsInteger() {
			return object.ConvertToInt(obj.Type(), obj.(object.Integer).Integer()-int64(1))
		} else if obj.Type().IsFloat() {
			return object.ConvertToFloat(obj.Type(), -obj.(object.Float).Float()-float64(1))
		} else {
			return object.NewError("operator -- not defined on %s", obj.Type())
		}
	}
	return nil
}

func doBinaryExpr(op code.Opcode, left, right object.Object) object.Object {
	switch left.Type() {
	case object.SINGLE_RETURN_OBJ:
		left = left.(*object.SingleReturn).Value
	case object.MAP_EXIST_OBJ:
		left = left.(*object.MapExist).Value
	}

	switch right.Type() {
	case object.SINGLE_RETURN_OBJ:
		right = right.(*object.SingleReturn).Value
	case object.MAP_EXIST_OBJ:
		right = right.(*object.MapExist).Value
	}

	if left.Type() != right.Type() {
		return object.NewError("Binary mismatched types %s and %s", left.Type(), right.Type())
	}

	switch {
	case left.Type().IsInteger():
		return doIntegerBinaryExpr(op, left, right)
	case left.Type().IsFloat():
		return doFloatBinaryExpr(op, left, right)
	case left.Type() == object.BOOLEAN_OBJ:
		return doBooleanBinaryExpr(op, left, right)
	case left.Type() == object.STRING_OBJ:
		return doStringBinaryExpr(op, left, right)
	case left.Type() == object.SINGLE_RETURN_OBJ:
		return doBinaryExpr(op, left.(*object.SingleReturn).Value, right.(*object.SingleReturn).Value)
	default:
		return object.NewError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func doIntegerBinaryExpr(op code.Opcode, left, right object.Object) object.Object {
	lv := left.(object.Integer).Integer()
	rv := right.(object.Integer).Integer()
	switch op {
	case code.OpADD:
		return object.ConvertToInt(left.Type(), lv+rv)
	case code.OpSUB:
		return object.ConvertToInt(left.Type(), lv-rv)
	case code.OpMUL:
		return object.ConvertToInt(left.Type(), lv*rv)
	case code.OpQUO:
		return object.ConvertToInt(left.Type(), lv/rv)
	case code.OpREM:
		return object.ConvertToInt(left.Type(), lv%rv)
	case code.OpAND:
		return object.ConvertToInt(left.Type(), lv&rv)
	case code.OpOR:
		return object.ConvertToInt(left.Type(), lv|rv)
	case code.OpXOR:
		return object.ConvertToInt(left.Type(), lv^rv)
	case code.OpSHL:
		return object.ConvertToInt(left.Type(), lv<<rv)
	case code.OpSHR:
		return object.ConvertToInt(left.Type(), lv>>rv)
	case code.OpAND_NOT:
		return object.ConvertToInt(left.Type(), lv&^rv)
	case code.OpEQL:
		return object.ConvertToBoolean(lv == rv)
	case code.OpLSS:
		return object.ConvertToBoolean(lv < rv)
	case code.OpGTR:
		return object.ConvertToBoolean(lv > rv)
	case code.OpNEQ:
		return object.ConvertToBoolean(lv != rv)
	case code.OpLEQ:
		return object.ConvertToBoolean(lv <= rv)
	case code.OpGEQ:
		return object.ConvertToBoolean(lv >= rv)
	default:
		return object.NewError("the operator %s is not defined on %s", op, left.Type())
	}
}

func doFloatBinaryExpr(op code.Opcode, left, right object.Object) object.Object {
	lv := left.(object.Float).Float()
	rv := right.(object.Float).Float()
	switch op {
	case code.OpADD:
		return object.ConvertToFloat(left.Type(), lv+rv)
	case code.OpSUB:
		return object.ConvertToFloat(left.Type(), lv-rv)
	case code.OpMUL:
		return object.ConvertToFloat(left.Type(), lv*rv)
	case code.OpQUO:
		return object.ConvertToFloat(left.Type(), lv/rv)
	case code.OpEQL:
		return object.ConvertToBoolean(lv == rv)
	case code.OpLSS:
		return object.ConvertToBoolean(lv < rv)
	case code.OpGTR:
		return object.ConvertToBoolean(lv > rv)
	case code.OpNEQ:
		return object.ConvertToBoolean(lv != rv)
	case code.OpLEQ:
		return object.ConvertToBoolean(lv <= rv)
	case code.OpGEQ:
		return object.ConvertToBoolean(lv >= rv)
	default:
		return object.NewError("the operator %s is not defined on %s", op, left.Type())
	}
}

func doBooleanBinaryExpr(op code.Opcode, left, right object.Object) object.Object {
	lv := left.(*object.Boolean).Value
	rv := right.(*object.Boolean).Value
	switch op {
	case code.OpLAND:
		return object.ConvertToBoolean(lv && rv)
	case code.OpLOR:
		return object.ConvertToBoolean(lv || rv)
	case code.OpEQL:
		return object.ConvertToBoolean(lv == rv)
	case code.OpNEQ:
		return object.ConvertToBoolean(lv != rv)
	default:
		return object.NewError("the operator %s is not defined on %s", op, left.Type())
	}
}

func doStringBinaryExpr(op code.Opcode, left, right object.Object) object.Object {
	lv := left.(*object.String).Value
	rv := right.(*object.String).Value
	if op == code.OpADD {
		return &object.String{Value: lv + rv}
	} else {
		cp := strings.Compare(lv, rv)
		switch op {
		case code.OpEQL:
			if cp == 0 {
				return object.TRUE
			} else {
				return object.FALSE
			}
		case code.OpNEQ:
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

func extractData(obj object.Object) (newValue, sourceObj object.Object, needPop bool) {
	switch tobj := obj.(type) {
	case *object.MapExist:
		if !tobj.SkipValue {
			tobj.SkipValue = true
			newValue = tobj.Value
		} else {
			newValue = &object.Boolean{Value: tobj.Exist}
			needPop = true
		}
		sourceObj = tobj
	case *object.MultiReturn:
		newValue = tobj.Values[0]

		tobj.Values = tobj.Values[1:]
		if len(tobj.Values) == 0 {
			needPop = true
		}
		sourceObj = tobj
	case *object.SingleReturn:
		newValue = tobj.Value
		needPop = true
	default:
		newValue = tobj
		needPop = true
	}
	return
}
