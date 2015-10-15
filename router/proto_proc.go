package main

import (
	"flag"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type ProtoProc struct {
	Router *Router
}

func NewProtoProc(r *Router) *ProtoProc {
	return &ProtoProc{
		Router: r,
	}
}

func (self *ProtoProc) procSubscribeChannel(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSubscribeChannel")
	var err error

	if len(cmd.GetArgs()) < 1 {
		return err
	}

	router := cmd.GetArgs()[0]

	self.Router.brotherServerMutex.Lock()
	self.Router.brotherServerMap[router] = session
	self.Router.brotherServerMutex.Unlock()

	return err
}

func (self *ProtoProc) procRouteMsg(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMsg")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_MSG_CMD_ARGS_NUM {
		return err
	}

	targetServer := cmd.GetArgs()[0]

	if self.Router.msgServerClientMap[targetServer] != nil {
		err = self.Router.msgServerClientMap[targetServer].Send(cmd)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	} else {
		router := self.getRouterFromMsgServer(targetServer)
		if self.Router.brotherServerMap[router] != nil {
			err = self.Router.brotherServerMap[router].Send(cmd)
			if err != nil {
				log.Error("error:", err)
				return err
			}
		}
	}

	return err
}

//根据msgServer找到router
func (self *ProtoProc) getRouterFromMsgServer(msgServer string) string {
	var router string

	for k, v := range self.Router.otherMsgServerMap {
		if common.InArray(v, msgServer) {
			return k
		}
	}

	return router
}
