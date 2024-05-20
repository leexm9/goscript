package vm

import (
	"errors"
	"goscript/compiler"
	"goscript/object"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // 始终指向栈中的下一个空槽位

	globals    []object.Object
	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, bytecode.SymbolTable.NumDefinitions)
	mainFrame.IsMain = true

	globals := make([]object.Object, GlobalsSize)
	j := 0
	for i := 0; i < len(bytecode.Constants); i++ {
		fn, ok := bytecode.Constants[i].(*object.CompiledFunction)
		if ok && fn.Name != "" {
			globals[j] = &object.Closure{Fn: fn, Free: make([]object.Object, 0)}
			j++
		}
	}

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants:  bytecode.Constants,
		stack:      make([]object.Object, StackSize),
		sp:         0,
		globals:    globals,
		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp > StackSize {
		return errors.New("stack overflow")
	}
	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}
