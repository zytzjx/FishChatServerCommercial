package main

import (
	// "goProject/protocol"
	"sync"
)

var MsgServerInfoMutex sync.Mutex

type InfoData map[string]interface{}

var NewestMsgServerInfoData InfoData

func InitContainer() {
	NewestMsgServerInfoData = make(InfoData)
}
