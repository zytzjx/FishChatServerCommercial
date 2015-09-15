package main

import (
	"encoding/json"
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
		err, router := self.getRouterFromMsgServer(storeSession.MsgServerAddr)
		if err != nil {
			log.Info("Not find servers")
		} else {
			if self.Router.brotherServerMap[router] != nil {
				err = self.Router.brotherServerMap[router].Send(cmd)
				if err != nil {
					log.Error("error:", err)
					return err
				}
			}
		}
	}

	return err
}

// //新建群组
// func (self *ProtoProc) procCreateTopic(cmd protocol.Cmd, session *libnet.Session) error {
// 	log.Info("procCreateTopic")
// 	topicName := cmd.GetArgs()[0]
// 	serverAddr := cmd.GetAnyData().(string)
// 	self.Router.topicServerMap[topicName] = serverAddr

// 	return nil
// }

// //Note: router do not process topic
// //加入群组
// func (self *ProtoProc) procJoinTopic(cmd protocol.Cmd, session *libnet.Session) error {
// 	log.Info("procJoinTopic")

// 	return nil
// }

//Note: router do not process topic
//发送群组信息
func (self *ProtoProc) procSendMsgTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMsgTopic")
	var err error
	if len(cmd.GetArgs()) < protocol.ROUTE_MESSAGE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return err
	}

	getTargetListStr := cmd.GetArgs()[4]
	getTargetListByte := []byte(getTargetListStr)

	var Clients []mongo_store.SessionStoreData
	json.Unmarshal(getTargetListByte, &Clients)

	if self.Router.msgServerClientMap[Clients[0].MsgServerAddr] != nil {
		cmd.ChangeCmdName(protocol.ROUTE_MESSAGE_TOPIC_CMD)
		err = self.Router.msgServerClientMap[Clients[0].MsgServerAddr].Send(cmd)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	} else {
		err, router := self.getRouterFromMsgServer(Clients[0].MsgServerAddr)
		if err != nil {
			log.Info("Not find servers")
		} else {
			if self.Router.brotherServerMap[router] != nil {
				err = self.Router.brotherServerMap[router].Send(cmd)
				if err != nil {
					log.Error("error:", err)
					return err
				}
			}
		}
	}

	return nil
}

//发送切换MsgServer命令
func (self *ProtoProc) procChangeMessageServer(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procChangeMessageServer")
	var err error
	if len(cmd.GetArgs()) < protocol.ROUTE_CHANGE_MESSAGE_SERVER_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return err
	}

	msgServerAddr := cmd.GetArgs()[1]

	self.Router.readMutex.Lock()
	defer self.Router.readMutex.Unlock()

	if self.Router.msgServerClientMap[msgServerAddr] != nil {
		cmd.ChangeCmdName(protocol.ROUTE_CHANGE_MESSAGE_SERVER_CMD)
		err = self.Router.msgServerClientMap[msgServerAddr].Send(cmd)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	} else {
		err, router := self.getRouterFromMsgServer(msgServerAddr)
		if err != nil {
			log.Info("Not find servers")
		} else {
			if self.Router.brotherServerMap[router] != nil {
				err = self.Router.brotherServerMap[router].Send(cmd)
				if err != nil {
					log.Error("error:", err)
					return err
				}
			}
		}
	}

	return err
}

//发送请求信息
func (self *ProtoProc) procSendAskMsg(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendAskMsg")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_ASK_CMD_ARGS_NUM {
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

	// cmd.ChangeCmdName(protocol.ROUTE_ASK_CMD)
	// err = self.Router.msgServerClientMap[storeSession.MsgServerAddr].Send(cmd)
	// if err != nil {
	// 	log.Error("error:", err)
	// 	return err
	// }

	if self.Router.msgServerClientMap[storeSession.MsgServerAddr] != nil {
		cmd.ChangeCmdName(protocol.ROUTE_ASK_CMD)
		err = self.Router.msgServerClientMap[storeSession.MsgServerAddr].Send(cmd)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	} else {
		err, router := self.getRouterFromMsgServer(storeSession.MsgServerAddr)
		if err != nil {
			log.Info("Not find servers")
		} else {
			if self.Router.brotherServerMap[router] != nil {
				err = self.Router.brotherServerMap[router].Send(cmd)
				if err != nil {
					log.Error("error:", err)
					return err
				}
			}
		}
	}

	return err
}

//根据msgServer找到router
func (self *ProtoProc) getRouterFromMsgServer(msgServer string) (error, string) {
	var router string
	var err error

	for k, v := range self.Router.otherMsgServerMap {
		if common.InArray(v, msgServer) {
			return nil, k
		}
	}

	return err, router
}
