package main

import (
	"encoding/json"
	"flag"
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

func (self *ProtoProc) procSendMsgP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMsgP2P")
	var err error
	send2Msg := cmd.GetArgs()[0]
	send2ID := cmd.GetArgs()[1]
	log.Info(send2Msg)
	self.Router.readMutex.Lock()
	defer self.Router.readMutex.Unlock()

	storeSession, err := self.Router.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME,
		mongo_store.CLIENT_INFO_COLLECTION, send2ID)
	if err != nil {
		log.Error("error:", err)
		return err
	}
	log.Info(storeSession.MsgServerAddr)

	cmd.ChangeCmdName(protocol.ROUTE_MESSAGE_P2P_CMD)

	err = self.Router.msgServerClientMap[storeSession.MsgServerAddr].Send(cmd)
	if err != nil {
		log.Error("error:", err)
		return err
	}

	return err
}

//新建群组
func (self *ProtoProc) procCreateTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procCreateTopic")
	topicName := cmd.GetArgs()[0]
	serverAddr := cmd.GetAnyData().(string)
	self.Router.topicServerMap[topicName] = serverAddr

	return nil
}

//Note: router do not process topic
//加入群组
func (self *ProtoProc) procJoinTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procJoinTopic")

	return nil
}

//Note: router do not process topic
//发送群组信息
func (self *ProtoProc) procSendMsgTopic(cmd *protocol.CmdInternal, session *libnet.Session) error {
	log.Info("procSendMsgTopic")
	log.Info(cmd)
	cmd.ChangeCmdName(protocol.ROUTE_MESSAGE_TOPIC_CMD)
	log.Info("----------------------------------------------------------")
	log.Info(cmd)

	// self.Router.readMutex.Lock()
	// defer self.Router.readMutex.Unlock()
	// {SEND_MESSAGE_TOPIC [hellootherserver talks cc 2015-08-08 15:45:25 -0400 EDT [{"ClientID":"aa","ClientAddr":"192.168.60.101:34265","MsgServerAddr":"127.0.0.1:19001","Alive":true},{"ClientID":"bb","ClientAddr":"192.168.60.101:34267","MsgServerAddr":"127.0.0.1:19001","Alive":true}]346b68ba-1ca7-4fcd-bbc4-ea93034e7e8e] <nil>}
	getTargetListStr := cmd.GetArgs()[4]
	getTargetListByte := []byte(getTargetListStr)

	var Clients []mongo_store.SessionStoreData
	json.Unmarshal(getTargetListByte, &Clients)

	err := self.Router.msgServerClientMap[Clients[0].MsgServerAddr].Send(cmd)
	if err != nil {
		log.Error("error:", err)
		return err
	}

	return nil
}
