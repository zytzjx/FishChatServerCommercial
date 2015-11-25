package main

import (
	"flag"
	// "goProject/common"
	"encoding/json"
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
	case protocol.TYPE_GATEWAY_SERVER:
	case protocol.TYPE_ROUTER_SERVER:
	case protocol.TYPE_MSG_SERVER_SERVER:
		var data protocol.MsgServerMonitorData
		have := false
		err = json.Unmarshal([]byte(cmd.Data), &data)
		if err != nil {
			log.Error("error:", err)
			return err
		}

		MsgServerInfoMutex.Lock()

		for k, v := range NewestMsgServerInfoData {
			if v.ServerAddr == session.State.(string) {
				v.Time = cmd.Time
				v.SessionNum = data.SessionNum
				NewestMsgServerInfoData[k] = v

				have = true
			}
		}
		if have == false {
			NewestMsgServerInfoData = append(NewestMsgServerInfoData, MsgServerInfo{
				ServerAddr: session.State.(string),
				Time:       cmd.Time,
				SessionNum: data.SessionNum,
			})
		}

		MsgServerInfoMutex.Unlock()
	default:
	}

	return err
}
