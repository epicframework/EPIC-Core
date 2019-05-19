package main

import (
	"C"
	"github.com/sirupsen/logrus"
	"sync"
)

type LibSharedID int32
type LibSharedType int32

const (
	LibSharedTypeAPI    LibSharedType = 1
	LibSharedTypeList   LibSharedType = 2
	LibSharedTypeError  LibSharedType = 3
	LibSharedTypeResult LibSharedType = 4

	LibSharedTypeDebug LibSharedType = 500
)

type LibSharedItem struct {
	Type     LibSharedType
	Location *interface{}
}

var sharedItems map[LibSharedID]LibSharedItem
var sharedItemsNextId LibSharedID = 1

var globalLock *sync.Mutex

func AddSharedItem(typ LibSharedType, location *interface{}) LibSharedID {
	globalLock.Lock()
	defer globalLock.Unlock()

	id := sharedItemsNextId
	sharedItemsNextId++
	sharedItems[id] = LibSharedItem{
		Type:     typ,
		Location: location,
	}

	return id
}

func GetSharedItem(typ LibSharedType, id LibSharedID) (*interface{}, bool) {
	globalLock.Lock()
	defer globalLock.Unlock()

	item, exists := sharedItems[id]
	if !exists {
		logrus.Debugf("could not find: type(%d) id(%d)", typ, id)
		return nil, false
	}

	if item.Type != typ {
		logrus.Debugf("type assertion failed: type(%d) id(%d)", typ, id)
		return item.Location, false
	}

	return item.Location, true
}

func AddSharedError(err error) LibSharedID {
	var ei *interface{}
	*ei = err
	return AddSharedItem(LibSharedTypeError, ei)
}

//export elconn_get_type
func elconn_get_type(datumID LibSharedID) LibSharedType {
	item, exists := sharedItems[datumID]
	if !exists {
		return -1
	}

	return item.Type
}
