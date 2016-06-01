package main

import (
	// "goProject/protocol"
	"sync"
)

var MsgServerInfoMutex sync.Mutex

// type InfoData []interface{}

type MsgServerInfo struct {
	ServerAddr string `json:"server_addr"`
	Time       int64  `json:"time"`
	SessionNum uint64 `json:"session_num"`
}

var NewestMsgServerInfoData []MsgServerInfo

func InitContainer() {
	// NewestMsgServerInfoData = make(InfoData)
}
