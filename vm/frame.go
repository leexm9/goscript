package vm

import (
	"goscript/code"
	"goscript/object"
)

type Frame struct {
	Cl          *object.Closure
	Ip          int
	BasePointer int
	IsLoop      bool
	IsMain      bool
}

func NewFrame(fn *object.Closure, basePointer int) *Frame {
	return &Frame{
		Cl:          fn,
		Ip:          -1,
		BasePointer: basePointer,
	}
}

func NewLoopFrame(fn *object.Closure, basePointer int) *Frame {
	return &Frame{
		Cl:          fn,
		Ip:          -1,
		BasePointer: basePointer,
		IsLoop:      true,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.Cl.Fn.Instructions
}

func (f *Frame) currentClosure() *object.Closure {
	return f.Cl
}
