package util

import (
	"github.com/Jeffail/tunny"
	"runtime"
)

var ncpu = runtime.NumCPU()

func NewNumCPUPool(fn func(interface{}) interface{}) *tunny.Pool {
	return tunny.NewFunc(ncpu, fn)
}

func NewDualNumCPUPool(fn func(interface{}) interface{}) *tunny.Pool {
	return tunny.NewFunc(2*ncpu, fn)
}

func NewQuadrupleNumCPUPool(fn func(interface{}) interface{}) *tunny.Pool {
	return tunny.NewFunc(4*ncpu, fn)
}

func NewNCpuPool(n int, fn func(interface{}) interface{}) *tunny.Pool {
	return tunny.NewFunc(n, fn)
}
