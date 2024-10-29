package vm

import (
	"karaoke/code"
	"karaoke/object"
)

type Frame struct {
	fn      *object.CompiledFunction
	ip      int
	basePtr int
}

func NewFrame(fun *object.CompiledFunction, baseptr int) *Frame {
	return &Frame{fn: fun, ip: -1, basePtr: baseptr}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
