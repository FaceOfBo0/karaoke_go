package vm

import (
	"karaoke/code"
	"karaoke/object"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
}

func NewFrame(fun *object.CompiledFunction) *Frame {
	return &Frame{fn: fun, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
