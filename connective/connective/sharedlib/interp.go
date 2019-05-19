package main

import (
	"C"
	"github.com/KernelDeimos/anything-gos/interp_a"
	"github.com/sirupsen/logrus"
)

//export elconn_make_interpreter
func elconn_make_interpreter() LibSharedID {
	evaluator := interp_a.InterpreterFactoryA{}.MakeExec()

	// Evaluate function must by in an empty interface to take reference
	var op interface{}
	op = interp_a.Operation(evaluator.OpEvaluate)

	// Add data structure factory
	evaluator.AddOperation("@", makeDataStructureFactory())

	// Add plugin factory
	evaluator.AddOperation("include", makePluginFactory())

	// Share evaluate function with library caller
	id := AddSharedItem(LibSharedTypeAPI, &op)
	return id
}

//export elconn_call
func elconn_call(inOpID LibSharedID, inListID LibSharedID) LibSharedID {
	// Obtain caller inputs
	opInterface, okay := GetSharedItem(LibSharedTypeAPI, inOpID)
	if !okay {
		logrus.Error("call operation: invalid value")
		return 0
	}
	listInterface, okay := GetSharedItem(LibSharedTypeList, inListID)
	if !okay {
		logrus.Error("call list: invalid value")
		return 0
	}

	// Dereference caller inputs
	inOp := (*opInterface).(interp_a.Operation)
	inList := (*listInterface).([]interface{})

	// Run the specified operation with the specified input list
	result, err := inOp(inList)
	if err != nil {
		logrus.Error(err)
		return 0
	}

	// Share result list with caller
	var resultInterface interface{}
	resultInterface = result
	id := AddSharedItem(LibSharedTypeList, &resultInterface)
	return id
}

//export elconn_link
func elconn_link(name *C.char, srcID, dstID LibSharedID) {
	// Obtain caller inputs
	opSrcPtr, okay := GetSharedItem(LibSharedTypeAPI, srcID)
	if !okay {
		logrus.Error("call operation: invalid value")
		return
	}
	opDstPtr, okay := GetSharedItem(LibSharedTypeAPI, dstID)
	if !okay {
		logrus.Error("call operation: invalid value")
		return
	}
	nameStr := C.GoString(name)

	opSrc := (*opSrcPtr).(interp_a.Operation)
	opDst := (*opDstPtr).(interp_a.Operation)

	r, e := opDst([]interface{}{"format", "hi"})
	logrus.Debug(r)
	logrus.Warn(e)

	result, err := opSrc([]interface{}{"evaluator:", nameStr, opDst})
	if err != nil {
		logrus.Error(err)
		logrus.Debug(result)
	}
}
