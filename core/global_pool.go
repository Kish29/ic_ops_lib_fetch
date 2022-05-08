package core

import "github.com/Kish29/ic_ops_lib_fetch/pool"

var (
	GlobalPool  = pool.New(2048)
	GlobalPool2 = pool.New(2048)
)
