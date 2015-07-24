

package main

import (
	"flag"
	"goProject/log"
	"goProject/libnet"
	"goProject/base"
	"goProject/common"
	"goProject/protocol"
	"goProject/storage/mongo_store"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type ProtoProc struct {
	msgServer    *MsgServer
}

func NewProtoProc(msgServer *MsgServer) *ProtoProc {
	return &ProtoProc {
		msgServer : msgServer,
	}
}

func (self *ProtoProc)procSubscribeChannel(cmd protocol.Cmd, session *libnet.Session) {
	log.Info("procSubscribeChannel")
	channelName := cmd.GetArgs()[0]
	cUUID := cmd.GetArgs()[1]
	log.Info(channelName)
	if self.msgServer.channels[channelName] != nil {
		//fixme
		session.EnableAsyncSend(1024)
		self.msgServer.channels[channelName].Channel.Join(session)
		self.msgServer.channels[channelName].ClientIDlist = append(self.msgServer.channels[channelName].ClientIDlist, cUUID)
	} else {
		log.Warning(channelName + " is not exist")
	}

	log.Info(self.msgServer.channels)
}

func (self *ProtoProc)procPing(cmd protocol.Cmd, session *libnet.Session) error {
	//log.Info("procPing")
	cid := session.State.(*base.SessionState).ClientID
	self.msgServer.scanSessionMutex.Lock()
	defer self.msgServer.scanSessionMutex.Unlock()
	self.msgServer.sessions[cid].State.(*base.SessionState).Alive = true
	
	return nil
}

//func (self *ProtoProc)procOfflineMsg(session *libnet.Session, ID string) error {
//	var err error
//	exist, err := self.msgServer.offlineMsgCache.IsKeyExist(ID)
//	if exist.(int64) == 0 {
//		return err
//	} else {
//		omrd, err := common.GetOfflineMsgFromOwnerName(self.msgServer.offlineMsgCache, ID)
//		if err != nil {
//			log.Error(err.Error())
//			return err
//		}
//		for  _, v := range omrd.MsgList {
//			resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_P2P_CMD)
//			resp.AddArg(v.Msg)
//			resp.AddArg(v.FromID)
//			resp.AddArg(v.Uuid)
			
//			if self.msgServer.sessions[ID] != nil {
//				self.msgServer.sessions[ID].Send(libnet.Json(resp))
//				if err != nil {
//					log.Error(err.Error())
//					return err
//				}
//			} 
//		}
		
//		omrd.ClearMsg()
//		self.msgServer.offlineMsgCache.Set(omrd)
//	}
	
//	return err
//}

func (self *ProtoProc)procClientID(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procClientID")
	cid := cmd.GetArgs()[0]
	
	self.msgServer.sessions[cid] = session
	self.msgServer.sessions[cid].State = base.NewSessionState(cmd.GetArgs()[0], true, 
		session.Conn().RemoteAddr().String(), self.msgServer.cfg.LocalIP)
		
	sessionStoreData := mongo_store.SessionStoreData{cid, session.Conn().RemoteAddr().String(), 
		self.msgServer.cfg.LocalIP, true}
	
	// update login info	
	self.msgServer.mongoStore.Update(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
	
	

	return nil
}

func (self *ProtoProc)procSendMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMessageP2P")
	var err error
	send2ID := cmd.GetArgs()[0]
	send2Msg := cmd.GetArgs()[1]
	fromID := cmd.GetArgs()[2]
	
	uuid := common.NewV4().String()
	
	log.Info("uuid : ", uuid)
	
	uuidTmpMap := make(map[string]bool)
	uuidTmpMap[uuid] = false
	
	self.msgServer.p2pAckStatus[fromID] = uuidTmpMap
	
//	if self.msgServer.sessions[send2ID] == nil {
//		//offline
//		log.Info(send2ID + " | is offline")
//		return nil
//	}
	
	if self.msgServer.sessions[send2ID] != nil {
		log.Info("in the same server")
		resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_P2P_CMD)
		resp.AddArg(send2Msg)
		resp.AddArg(fromID)
		// add uuid
		resp.AddArg(uuid)
		
		if self.msgServer.sessions[send2ID] != nil {
			self.msgServer.sessions[send2ID].Send(resp)
			if err != nil {
				log.Error(err.Error())
			}
		} 
	} else {
		log.Info("Not in the same server")
		if self.msgServer.channels[protocol.SYSCTRL_SEND] != nil {
			//add uuid
			cmd.AddArg(uuid)
			err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(cmd)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	}
	
	return nil
}

func (self *ProtoProc)procRouteMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMessageP2P")
	var err error
	send2ID := cmd.GetArgs()[0]
	send2Msg := cmd.GetArgs()[1]
	fromID := cmd.GetArgs()[2]
	uuid := cmd.GetArgs()[3]
//	_, err = common.GetSessionFromCID(self.msgServer.sessionCache, send2ID)
//	if err != nil {
//		log.Warningf("no ID : %s", send2ID)
		
//		return err
//	}

	resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_P2P_CMD)
	resp.AddArg(send2Msg)
	resp.AddArg(fromID)
	// add uuid
	resp.AddArg(uuid)
	
	if self.msgServer.sessions[send2ID] != nil {
		self.msgServer.sessions[send2ID].Send(resp)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	return nil
}


// not a good idea
func (self *ProtoProc)procP2pAck(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procP2pAck")
	var err error
	clientID := cmd.GetArgs()[0]
	uuid := cmd.GetArgs()[1]
	self.msgServer.p2pAckMutex.Lock()
	defer self.msgServer.p2pAckMutex.Unlock()
	
	//self.msgServer.p2pAckStatus[clientID][uuid] = true
	
	m, ok := self.msgServer.p2pAckStatus[clientID]
	if ok {
		m[uuid] = true
	}
	
	return err
}