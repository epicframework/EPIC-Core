package main

import (
	"C"
	"github.com/gin-gonic/gin"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	ModeNormal = 0
	ModeDebug  = 1
)

type DebugInfo struct {
	Message string
}

//export elconn_init
func elconn_init(mode int32) LibSharedID {
	var modeStr string

	switch mode {
	case ModeDebug:
		modeStr = "debug"
		logrus.SetLevel(logrus.DebugLevel)
	case ModeNormal:
		modeStr = "normal"
		gin.SetMode(gin.ReleaseMode)
	}

	globalLock = &sync.Mutex{}
	sharedItems = map[LibSharedID]LibSharedItem{}

	return logDebug("elconn_init() called in " + modeStr + " mode")
}

func main() {}
