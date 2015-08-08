package main

import (
	"encoding/json"
	"flag"
	"goProject/base"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	"time"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type ProtoProc struct {
	msgServer *MsgServer
}

func NewProtoProc(msgServer *MsgServer) *ProtoProc {
	return &ProtoProc{
		msgServer: msgServer,
	}
}

func (self *ProtoProc) procSubscribeChannel(cmd protocol.Cmd, session *libnet.Session) {
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

//ping命令解释
func (self *ProtoProc) procPing(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procPing")
	cid := session.State.(*base.SessionState).ClientID
	self.msgServer.scanSessionMutex.Lock()
	defer self.msgServer.scanSessionMutex.Unlock()
	self.msgServer.sessions[cid].State.(*base.SessionState).Alive = true

	self.msgServer.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, cid, true)

	return nil
}

//获取用户所有未读信息
func (self *ProtoProc) procOfflineMsg(session *libnet.Session, cid string) error {
	var err error
	log.Info("Read p2p offline message")
	//获取用户未读P2p信息
	err = self.procP2POfflineMsg(session, cid)
	if err != nil {
		return err
	}

	log.Info("Read topic offline message")
	//获取用户未读群组信息
	err = self.procTopicOfflineMsg(session, cid)
	if err != nil {
		return err
	}
	return err
}

//获取用户P2P未读信息
func (self *ProtoProc) procP2POfflineMsg(session *libnet.Session, cid string) error {
	var err error

	//从mongo读取信息
	recordData, err := self.msgServer.mongoStore.ReadP2PRecordMessage(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, cid)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	//把从数据库中取出的数据发送给Client
	for _, v := range recordData {
		resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_P2P_CMD)
		resp.AddArg(v.Content)
		resp.AddArg(v.FromID)
		resp.AddArg(time.Unix(v.Time, 0).String())
		resp.AddArg(v.UUID)

		if self.msgServer.sessions[cid] != nil {
			err = self.msgServer.sessions[cid].Send(resp)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	}

	//用户已经接收所有离线消息，现在把离线消息设为已读
	err = self.msgServer.mongoStore.MarkP2PRecordMessage(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, cid)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}

//获取用户群组未读信息
func (self *ProtoProc) procTopicOfflineMsg(session *libnet.Session, cid string) error {
	var err error

	//读取用户所有群组
	topics := self.msgServer.mongoStore.GetTopicsFromClientID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, cid)
	if topics == nil {
		log.Info("No topic")
		return err
	}

	topicsNameArr := make([]string, 0)
	for _, v := range topics {
		topicsNameArr = append(topicsNameArr, v.TopicID)
	}

	//读取用户所有群组未读信息
	recordData := self.msgServer.mongoStore.ReadTopicRecordMessage(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, cid, topicsNameArr)
	if recordData == nil {
		log.Info("No topic record data.")
		return err
	}

	for _, v := range recordData {
		resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_TOPIC_CMD)
		resp.AddArg(v.Content)
		resp.AddArg(v.FromID)
		resp.AddArg(v.ToID)
		resp.AddArg(time.Unix(v.Time, 0).String())
		resp.AddArg(v.UUID)

		if self.msgServer.sessions[cid] != nil {
			err = self.msgServer.sessions[cid].Send(resp)
			if err != nil {
				log.Error(err.Error())
				return err
			}

			//标记已读
			err = self.msgServer.mongoStore.MarkTopicRecordMessageFromObjectId(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, v.Id_, append(v.IsRead, cid))
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	}

	return err
}

//接收用户登录ID
func (self *ProtoProc) procClientID(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procClientID")
	cid := cmd.GetArgs()[0]

	self.msgServer.sessions[cid] = session
	self.msgServer.sessions[cid].State = base.NewSessionState(cmd.GetArgs()[0], true,
		session.Conn().RemoteAddr().String(), self.msgServer.cfg.LocalIP)

	sessionStoreData := mongo_store.SessionStoreData{cid, session.Conn().RemoteAddr().String(),
		self.msgServer.cfg.LocalIP, true}

	// update login info
	err := self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	//获取用户未读信息
	self.procOfflineMsg(session, cid)

	return nil
}

// 解析P2P信息
func (self *ProtoProc) procSendMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMessageP2P")
	var err error
	send2Msg := cmd.GetArgs()[0]
	send2ID := cmd.GetArgs()[1]
	fromID := cmd.GetArgs()[2]
	send2Time := time.Now().Unix()

	uuid := common.NewV4().String()
	uuidTmpMap := make(map[string]bool)
	uuidTmpMap[uuid] = false
	self.msgServer.p2pAckStatus[fromID] = uuidTmpMap

	if self.msgServer.sessions[send2ID] != nil {
		log.Info("In the same server")
		resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_P2P_CMD)
		resp.AddArg(send2Msg)
		resp.AddArg(fromID)
		resp.AddArg(string(send2Time))
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
			}
		}
	}

	alive, err := self.msgServer.mongoStore.IsSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, send2ID)

	if err != nil {
		log.Error(err.Error())
		return err
	} else {

		is_online := false
		if alive {
			is_online = true
		}

		//保存消息到mongodb中
		data := mongo_store.P2PRecordMessageData{fromID, send2ID, send2Msg, send2Time, uuid, is_online}
		err_record := self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, &data)
		if err_record != nil {
			log.Error(err_record.Error())
			return err_record
		} else {
			if !alive {
				log.Info(send2ID + " | is offline")
				log.Info(fromID + "send to " + send2ID + " of the data has been saved to the database.")
				return nil
			}
		}

	}

	return err
}

//解析Router P2P信息
func (self *ProtoProc) procRouteMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMessageP2P")
	var err error
	send2Msg := cmd.GetArgs()[0]
	send2ID := cmd.GetArgs()[1]
	fromID := cmd.GetArgs()[2]
	uuid := cmd.GetArgs()[3]

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
func (self *ProtoProc) procP2pAck(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procP2pAck")
	var err error
	clientID := cmd.GetArgs()[0]
	uuid := cmd.GetArgs()[1]
	self.msgServer.p2pAckMutex.Lock()
	defer self.msgServer.p2pAckMutex.Unlock()

	m, ok := self.msgServer.p2pAckStatus[clientID]
	if ok {
		m[uuid] = true
	}

	return err
}

// 增加Topic
func (self *ProtoProc) procCreateTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procCreateTopic")
	var err error

	//群组ID
	topicId := cmd.GetArgs()[0]
	founderId := cmd.GetArgs()[1]

	// 如果群组不存在,才添加群组
	if result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId); result == nil {
		//要存入数据库的数据
		ClientsID := []string{founderId}
		TopicStoreData := mongo_store.TopicStoreData{topicId, self.msgServer.cfg.LocalIP, founderId, ClientsID, true}

		err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}

	return nil
}

//加入Topic
func (self *ProtoProc) procJoinTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procJoinTopic")
	var err error

	//群组ID
	topicId := cmd.GetArgs()[0]
	clientId := cmd.GetArgs()[1]

	//如果群组存在,群组存在
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result != nil {
		users := result.ClientsID
		if !common.In_array(users, clientId) {
			users = append(users, clientId)
			TopicStoreData := mongo_store.TopicStoreData{topicId, self.msgServer.cfg.LocalIP, clientId, users, true}
			err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	}
	return nil
}

//处理Topic信息
func (self *ProtoProc) procSendMessageTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMessageTopic")

	var err error
	send2Msg := cmd.GetArgs()[0]
	topicId := cmd.GetArgs()[1]
	fromID := cmd.GetArgs()[2]
	send2Time := time.Now().Unix()

	uuid := common.NewV4().String()

	uuidTmpMap := make(map[string]bool)
	uuidTmpMap[uuid] = false

	self.msgServer.p2pAckStatus[fromID] = uuidTmpMap

	//获取Topic的信息
	topicResult := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if topicResult == nil {
		log.Info("no topic in db")
		return err
	}

	//判断用户是否属于该Topic
	if !common.In_array(topicResult.ClientsID, fromID) {
		log.Info(fromID + " don't belong to the " + topicId)
		return err
	}

	//获取群组成员信息
	msgResult := self.msgServer.mongoStore.GetOnlineClientsFromIds(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, topicResult.ClientsID)

	if msgResult == nil {
		log.Info("no client list")
	}

	var clientGroup map[string][]mongo_store.SessionStoreData
	clientGroup = make(map[string][]mongo_store.SessionStoreData)
	onlineUsers := make([]string, 0) //所有在线用户，后面保存信息记录时全部设置成已读

	for _, v := range msgResult {
		if clientGroup[v.MsgServerAddr] != nil {
			clientGroup[v.MsgServerAddr] = append(clientGroup[v.MsgServerAddr], *v)
		} else {
			clientGroup[v.MsgServerAddr] = make([]mongo_store.SessionStoreData, 1)
			clientGroup[v.MsgServerAddr][0] = *v
		}
		onlineUsers = append(onlineUsers, v.ClientID)
	}

	//直接到客户端的信息
	resp := protocol.NewCmdSimple(protocol.RESP_MESSAGE_TOPIC_CMD)
	resp.AddArg(send2Msg)
	resp.AddArg(fromID)
	resp.AddArg(topicId)
	resp.AddArg(time.Unix(send2Time, 0).String())
	resp.AddArg(uuid)

	// map[127.0.0.1:19001:[{aa 192.168.60.101:57826 127.0.0.1:19001 false} {bb 192.168.60.101:57829 127.0.0.1:19001 true} {cc 192.168.60.101:57845 127.0.0.1:19001 true}]]
	//&{SEND_MESSAGE_TOPIC [hello topics aa]}
	for k, v := range clientGroup {
		if k == self.msgServer.cfg.LocalIP {
			log.Info("In the same server")
			for _, client := range v {
				if self.msgServer.sessions[client.ClientID] != nil {
					err = self.msgServer.sessions[client.ClientID].Send(resp)
					if err != nil {
						log.Error(err.Error())
					}
				}
			}
		} else {
			log.Info("Not in the same server")
			json, _ := json.Marshal(v)

			//router转发的信息
			routerCmd := protocol.NewCmdSimple(protocol.SEND_MESSAGE_TOPIC_CMD)
			routerCmd.AddArg(send2Msg)
			routerCmd.AddArg(topicId)
			routerCmd.AddArg(fromID)
			routerCmd.AddArg(time.Unix(send2Time, 0).String())
			routerCmd.AddArg(string(json))
			routerCmd.AddArg(uuid)

			if self.msgServer.channels[protocol.SYSCTRL_SEND] != nil {
				err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(routerCmd)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}

	//保存消息到mongodb中
	data := mongo_store.TopicRecordMessageData{fromID, topicId, send2Msg, send2Time, uuid, onlineUsers}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, &data)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info(fromID + " send to " + topicId + " of the data has been saved to the database.")

	return err
}

//解析Router Client信息
func (self *ProtoProc) procRouteMessageTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMessageTopic")

	var err error
	args := cmd.GetArgs()
	send2Msg, topicId, fromID, send2Time, getTargetListStr, uuid := args[0], args[1], args[2], args[3], args[4], args[5]

	getTargetListByte := []byte(getTargetListStr)
	var Clients []mongo_store.SessionStoreData
	json.Unmarshal(getTargetListByte, &Clients)

	for _, v := range Clients {
		//router转发的信息
		newCmd := protocol.NewCmdSimple(protocol.SEND_MESSAGE_TOPIC_CMD)
		newCmd.AddArg(send2Msg)
		newCmd.AddArg(topicId)
		newCmd.AddArg(fromID)
		newCmd.AddArg(send2Time)
		newCmd.AddArg(uuid)

		if self.msgServer.sessions[v.ClientID] != nil {
			err = self.msgServer.sessions[v.ClientID].Send(newCmd)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	return nil
}
