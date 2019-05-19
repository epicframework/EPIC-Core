package main

import (
	"C"

	"github.com/sirupsen/logrus"
)

func logDebug(message string) LibSharedID {
	datum := DebugInfo{message}
	logrus.Debug(datum.Message)
	itype := interface{}(datum)

	return AddSharedItem(LibSharedTypeDebug, &itype)
}

//export elconn_display_info
func elconn_display_info(debugID LibSharedID) {
	item, ok := GetSharedItem(LibSharedTypeDebug, debugID)
	if ok {
		debug := (*item).(DebugInfo)
		logrus.Info(debug.Message)
	}
}
