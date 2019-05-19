package main

import (
	"C"

	"encoding/json"
	"github.com/sirupsen/logrus"

	"github.com/KernelDeimos/anything-gos/interp_a"
	"github.com/KernelDeimos/gottagofast/toolparse"
)

//export elconn_list_from_json
func elconn_list_from_json(jsonInputC *C.char) LibSharedID {
	jsonInput := C.GoString(jsonInputC)

	outList := []interface{}{}
	err := json.Unmarshal([]byte(jsonInput), &outList)
	if err != nil {
		logrus.Error(err)
		return 0
	}

	var outListInterface interface{}
	outListInterface = outList

	id := AddSharedItem(LibSharedTypeList, &outListInterface)
	return id
}

//export elconn_list_from_text
func elconn_list_from_text(textInputC *C.char) LibSharedID {
	textInput := C.GoString(textInputC)

	logrus.Info(textInput)

	outList, err := toolparse.ParseListSimple(textInput)
	if err != nil {
		logrus.Error(err)
		return 0
	}

	var outListInterface interface{}
	outListInterface = outList

	id := AddSharedItem(LibSharedTypeList, &outListInterface)
	return id
}

//export elconn_list_to_json
func elconn_list_to_json(listID LibSharedID) *C.char {
	listInterface, okay := GetSharedItem(LibSharedTypeList, listID)
	if !okay {
		logrus.Error("could not print list: invalid value")
		return C.CString("Err")
	}

	list := (*listInterface).([]interface{})

	// Remove invalid values from list
	for i, val := range list {
		switch val.(type) {
		case interp_a.Operation:
			list[i] = "__operation__"
		}
	}

	result, err := json.Marshal(list)
	if err != nil {
		logrus.Error(err)
		return C.CString("Err")
	}

	return C.CString(string(result))
}

//export elconn_list_strfirst
func elconn_list_strfirst(listID LibSharedID) *C.char {
	listInterface, okay := GetSharedItem(LibSharedTypeList, listID)
	if !okay {
		logrus.Error("could not print list: invalid value")
		return C.CString("Err")
	}

	listyList := (*listInterface).([]interface{})

	if len(listyList) < 1 {
		return C.CString("Err")
	}

	firstItem, ok := listyList[0].(string)
	if !ok {
		return C.CString("First item in list must be string")
	}

	return C.CString(firstItem)
}

//export elconn_list_print
func elconn_list_print(listID LibSharedID) int32 {
	listInterface, okay := GetSharedItem(LibSharedTypeList, listID)
	if !okay {
		logrus.Error("could not print list: invalid value")
		return -1
	}

	result, err := json.Marshal(*listInterface)
	if err != nil {
		logrus.Error(err)
		return -1
	}

	logrus.Info(string(result))
	return 0
}
