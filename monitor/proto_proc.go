package main

import (
	"flag"
	// "goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type ProtoProc struct {
	Monitor *Monitor
}

func NewProtoProc(r *Monitor) *ProtoProc {
	return &ProtoProc{
		Monitor: r,
	}
}

// func (self *ProtoProc) procSubscribeChannel(cmd protocol.CmdMonitor, session *libnet.Session) error {
// 	log.Info("procSubscribeChannel")
// 	var err error

// 	if len(cmd.GetArgs()) < 1 {
// 		return err
// 	}

// 	return err
// }

func (self *ProtoProc) procMonitorMsg(cmd protocol.CmdMonitor, session *libnet.Session) error {
	log.Info("procMonitorMsg")
	var err error

	switch cmd.ServerType {
	case protocol.TYPE_GATEWAY_SERVER_CMD:
	case protocol.TYPE_ROUTER_SERVER_CMD:
	case protocol.TYPE_MSG_SERVER_SERVER_CMD:
		MsgServerInfoMutex.Lock()
		NewestMsgServerInfoData[session.State.(string)] = cmd.Data
		MsgServerInfoMutex.Unlock()
	default:
	}

	return err
}
