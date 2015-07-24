

package main

import (
	"flag"
	"goProject/log"
	"goProject/libnet"
	"goProject/protocol"
	"goProject/storage/mongo_store"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type ProtoProc struct {
	Router   *Router
}

func NewProtoProc(r *Router) *ProtoProc {
	return &ProtoProc {
		Router : r,
	}
}

func (self *ProtoProc)procSendMsgP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMsgP2P")
	var err error
	send2ID := cmd.GetArgs()[0]
	send2Msg := cmd.GetArgs()[1]
	log.Info(send2Msg)
	self.Router.readMutex.Lock()
	defer self.Router.readMutex.Unlock()

	storeSession, err := self.Router.mongoStore.GetSessionFromCid(mongo_store.DATA_BASE_NAME, 
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

func (self *ProtoProc)procCreateTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procCreateTopic")
	topicName := cmd.GetArgs()[0]
	serverAddr := cmd.GetAnyData().(string)
	self.Router.topicServerMap[topicName] = serverAddr
	
	return nil
}

//Note: router do not process topic
func (self *ProtoProc)procJoinTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procJoinTopic")
	
	return nil
}

//Note: router do not process topic
func (self *ProtoProc)procSendMsgTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMsgTopic")
	//var err error
	//topicName := string(cmd.Args[0])
	//send2Msg := string(cmd.Args[1])

	
	return nil
}


