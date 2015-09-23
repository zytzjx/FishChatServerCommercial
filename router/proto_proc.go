package main

import (
	"flag"
	"goProject/common"
	"goProject/info"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
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
	self.Router.brotherServerMap[router] = session

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

func (self *ProtoProc) procSendMsgP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMsgP2P")
	var err error
	if len(cmd.GetArgs()) < protocol.ROUTE_MESSAGE_P2P_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return err
	}
	toID := cmd.GetArgs()[2]

	self.Router.readMutex.Lock()
	defer self.Router.readMutex.Unlock()

	storeSession, err := self.Router.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME,
		mongo_store.CLIENT_INFO_COLLECTION, toID)
	if err != nil {
		log.Error("error:", err)
		return err
	}

	if self.Router.msgServerClientMap[storeSession.MsgServerAddr] != nil {
		cmd.ChangeCmdName(protocol.ROUTE_MESSAGE_P2P_CMD)
		err = self.Router.msgServerClientMap[storeSession.MsgServerAddr].Send(cmd)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	} else {
		router := self.getRouterFromMsgServer(storeSession.MsgServerAddr)
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
