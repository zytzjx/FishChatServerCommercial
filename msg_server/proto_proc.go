package main

import (
	"encoding/json"
	"flag"
	"goProject/base"
	"goProject/common"
	"goProject/info"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	// "gopkg.in/mgo.v2/json"
	"strconv"
	"sync"
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

//router订阅请求
func (self *ProtoProc) procSubscribeChannel(cmd protocol.Cmd, session *libnet.Session) {
	log.Info("procSubscribeChannel")
	if len(cmd.GetArgs()) < 2 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return
	}
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
	// log.Info("procPing")

	if session.State == nil {
		self.respCmd(protocol.RESP_PONG_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}

	cid := session.State.(*base.SessionState).ClientID

	if self.msgServer.sessions[cid] == nil || self.msgServer.sessions[cid].State == nil {
		self.respCmd(protocol.RESP_PONG_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}

	self.msgServer.scanSessionMutex.Lock()
	self.msgServer.sessions[cid].State.(*base.SessionState).Alive = true
	self.msgServer.scanSessionMutex.Unlock()

	self.msgServer.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, cid, true)

	self.respCmd(protocol.RESP_PONG_CMD, session, cmd.GetReport(), true, "")
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

	log.Info("Read ask offline message")
	//获取用户未读群组信息
	err = self.procAskOfflineMsg(session, cid)
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
		receive := protocol.NewCmdResponse(NCommendMappedMap[v.MsgType].ReceiveCmd)
		receive.AddArg(v.Content)
		receive.AddArg(v.FromID)
		// strconv.FormatInt(v.Time, 10)
		receive.AddArg(strconv.FormatInt(v.Time, 10))
		receive.AddArg(v.UUID)

		//缓存uuid,等待ack
		ack := new(base.AckFrequency)
		ack.Frequency = 1
		ack.LastTime = time.Now().Unix()
		self.msgServer.p2pAckMap[v.UUID] = ack

		// time.Sleep(100)

		if self.msgServer.sessions[cid] != nil {
			err = self.msgServer.sessions[cid].Send(receive)
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
	var err error

	if len(cmd.GetArgs()) < protocol.SEND_CLIENT_ID_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		self.clientQuit(session)
		return nil
	}
	ClientID := cmd.GetArgs()[0]
	if len(ClientID) < 1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		self.clientQuit(session)
		return nil
	}

	//查找用户信息
	clientInfo, err := self.msgServer.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, ClientID)
	if err != nil {
		log.Error(err.Error())
	}
	if clientInfo != nil {
		if clientInfo.Alive == true {
			log.Info("User is logined in.")

			//如果用户已经登陆，就断开其他已经存在的连接
			if clientInfo.MsgServerAddr != self.msgServer.cfg.LocalIP {
				bMsg := protocol.NewCmdSimple(protocol.ROUTE_CHANGE_MESSAGE_SERVER_CMD)
				bMsg.AddArg(ClientID)

				temp, err := json.Marshal(bMsg)
				if err != nil {
					log.Error(err.Error())
					self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), false, info.ERROR)
					return err
				}

				routerMsg := protocol.NewCmdSimple(protocol.ROUTE_MSG_CMD)
				routerMsg.AddArg(clientInfo.MsgServerAddr)
				routerMsg.AddArg(string(temp))

				err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(routerMsg)
				if err != nil {
					log.Error(err.Error())
					self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), false, info.ERROR)
					return err
				}
			} else {
				if self.msgServer.sessions[ClientID] != nil {
					sMsg := protocol.NewCmdResponse(protocol.RESP_LOGOUT_CMD)
					sMsg.Message = info.YOU_HAVE_TO_RE_LOGIN
					err = self.msgServer.sessions[ClientID].Send(sMsg)
					if err != nil {
						log.Error(err.Error())
					}
					self.msgServer.sessions[ClientID].Close()
					self.deleteWithMutex(self.msgServer.sessions, ClientID, self.msgServer.scanSessionMutex)
				}
			}
		}

		sessionStoreData := mongo_store.SessionStoreData{ClientID, session.Conn().RemoteAddr().String(),
			self.msgServer.cfg.LocalIP, clientInfo.Friends, true}

		err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
		if err != nil {
			log.Error(err.Error())
			self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), false, info.ERROR)
			return err
		}
	} else {
		sessionStoreData := mongo_store.SessionStoreData{ClientID, session.Conn().RemoteAddr().String(),
			self.msgServer.cfg.LocalIP, []string{}, true}

		// update login info
		err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
		if err != nil {
			log.Error(err.Error())
			self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), false, info.ERROR)
			return err
		}
	}

	self.msgServer.sessions[ClientID] = session
	self.msgServer.sessions[ClientID].State = base.NewSessionState(ClientID, true,
		session.Conn().RemoteAddr().String(), self.msgServer.cfg.LocalIP)

	//获取用户未读信息
	go self.procOfflineMsg(session, ClientID)

	session.EnableAsyncSend(10)
	self.respCmd(protocol.RESP_CLIENT_ID_CMD, session, cmd.GetReport(), true, "")
	return err
}

//退出登录
func (self *ProtoProc) procLogout(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procLogout")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_LOGOUT_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	clientID := session.State.(*base.SessionState).ClientID
	if clientID != "" {
		// 标记用户离线
		err := self.msgServer.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientID, false)
		if err != nil {
			log.Error(err.Error())
			self.respCmd(protocol.RESP_LOGOUT_CMD, session, cmd.GetReport(), false, info.ERROR)
			return err
		}
	}

	self.respCmd(protocol.RESP_LOGOUT_CMD, session, cmd.GetReport(), true, "")
	self.clientQuit(session)
	return err
}

// 解析P2P信息
func (self *ProtoProc) procSendMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMessageP2P")
	var err error

	msgType := cmd.GetCmdName()

	if session.State == nil {
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_MESSAGE_P2P_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	fromID := session.State.(*base.SessionState).ClientID
	send2Msg := cmd.GetArgs()[0]
	send2ID := cmd.GetArgs()[1]
	send2Time := time.Now().Unix()
	uuid := common.NewV4().String()

	//保存消息到mongodb中
	data := mongo_store.P2PRecordMessageData{msgType, fromID, send2ID, send2Msg, send2Time, uuid, false}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, &data)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	if self.msgServer.sessions[send2ID] != nil {
		log.Info("In the same server")

		receive := protocol.NewCmdResponse(NCommendMappedMap[msgType].ReceiveCmd)
		receive.AddArg(send2Msg)
		receive.AddArg(fromID)
		receive.AddArg(strconv.FormatInt(send2Time, 10))
		receive.AddArg(uuid)

		if self.msgServer.sessions[send2ID] != nil {
			self.msgServer.sessions[send2ID].Send(receive)
			if err != nil {
				log.Error(err.Error())
				self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
				return err
			}
			//储存ACK，用来验证
			ack := new(base.AckFrequency)
			ack.Frequency = 1
			ack.LastTime = send2Time
			self.msgServer.p2pAckMap[uuid] = ack
		}
	} else {
		log.Info("Not in the same server")
		if self.msgServer.channels[protocol.SYSCTRL_SEND] != nil {

			storeSession, err := self.msgServer.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME,
				mongo_store.CLIENT_INFO_COLLECTION, send2ID)
			if err != nil {
				log.Error(err.Error())
				self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
				return err
			}
			rcmd := protocol.NewCmdSimple(NCommendMappedMap[msgType].RouterCmd)
			rcmd.AddArg(send2Msg)
			rcmd.AddArg(fromID)
			rcmd.AddArg(send2ID)
			rcmd.AddArg(strconv.FormatInt(send2Time, 10))
			rcmd.AddArg(uuid)

			temp, err := json.Marshal(rcmd)
			if err != nil {
				log.Error(err.Error())
				self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
				return err
			}
			routerMsg := protocol.NewCmdSimple(protocol.ROUTE_MSG_CMD)
			routerMsg.AddArg(storeSession.MsgServerAddr)
			routerMsg.AddArg(string(temp))

			err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(routerMsg)
			if err != nil {
				log.Error(err.Error())
				self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
				return err
			}
		}
	}

	self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), true, "")
	return err
}

// 解析P2P ACK信息
func (self *ProtoProc) procP2pAck(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procP2pAck")
	if len(cmd.GetArgs()) < protocol.P2P_ACK_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}
	log.Info(cmd)

	var err error
	uuid := cmd.GetArgs()[0]

	if self.msgServer.p2pAckMap[uuid] != nil {
		//InACK
		log.Info(uuid + " inACK list")
		//标记已读
		err = self.msgServer.mongoStore.MarkP2PRecordMessageFromUuid(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, uuid)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		self.deleteWithMutex(self.msgServer.p2pAckMap, uuid, self.msgServer.p2pAckMutex)
	}

	return err
}

//P2PACK超时重发处理
func (self *ProtoProc) procP2pTimeoutRetransmission() {
	// log.Info("procP2pTimeoutRetransmission")
	//储存ACK，用来验证

	for k, v := range self.msgServer.p2pAckMap {
		if v.Frequency >= protocol.P2P_ACK_FAILURES {
			log.Info(k + " is dead.")
			self.deleteWithMutex(self.msgServer.p2pAckMap, k, self.msgServer.p2pAckMutex)
			continue
		}
		if (time.Now().Unix() - v.LastTime) > protocol.P2P_ACK_TIMEOUT {
			//重设Ack
			self.msgServer.p2pAckMap[k].Frequency++
			self.msgServer.p2pAckMap[k].LastTime = time.Now().Unix()

			//从mongo读取信息
			recordData := self.msgServer.mongoStore.ReadP2PRecordMessageFromUuid(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, k)
			if recordData == nil {
				log.Info("No data.")
				continue
			}

			receive := protocol.NewCmdResponse(NCommendMappedMap[recordData.MsgType].ReceiveCmd)
			receive.AddArg(recordData.Content)
			receive.AddArg(recordData.FromID)
			// strconv.FormatInt(recordData.Time, 10)
			receive.AddArg(strconv.FormatInt(recordData.Time, 10))
			receive.AddArg(recordData.UUID)

			if self.msgServer.sessions[recordData.ToID] != nil {
				err := self.msgServer.sessions[recordData.ToID].Send(receive)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}
}

// 增加Topic
func (self *ProtoProc) procCreateTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procCreateTopic")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_CREATE_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_CREATE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_CREATE_TOPIC_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]
	founderId := session.State.(*base.SessionState).ClientID

	// 如果群组不存在,才添加群组
	if result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId); result != nil {
		self.respCmd(protocol.RESP_CREATE_TOPIC_CMD, session, cmd.GetReport(), false, info.TOPIC_ALREADY_EXISTS)
		return err
	}
	//要存入数据库的数据
	ClientsID := []string{founderId}
	TopicStoreData := mongo_store.TopicStoreData{topicId, founderId, ClientsID}

	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(protocol.RESP_CREATE_TOPIC_CMD, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	self.respCmd(protocol.RESP_CREATE_TOPIC_CMD, session, cmd.GetReport(), true, "")
	return nil
}

//加入Topic
func (self *ProtoProc) procJoinTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procJoinTopic")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_JOIN_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]
	clientId := session.State.(*base.SessionState).ClientID

	//如果群组存在,群组存在
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result == nil {
		log.Error(info.TOPIC_DOES_NOT_EXISTS)
		self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), false, info.TOPIC_DOES_NOT_EXISTS)
		return err
	}
	users := result.ClientsID
	if common.InArray(users, clientId) {
		log.Error(info.YOU_ARE_ALREADY_IN_THE_TOPIC)
		self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_ARE_ALREADY_IN_THE_TOPIC)
		return err
	}

	users = append(users, clientId)
	TopicStoreData := mongo_store.TopicStoreData{topicId, clientId, users}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), false, info.JOIN_TOPIC_FAILURE)
		return err
	}

	self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), true, "")

	// user string, timeNow int64
	err = self.msgServer.mongoStore.MarkTopicRecordMessageFromUserAndTime(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, clientId, time.Now().Unix(), topicId)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//邀请加入
func (self *ProtoProc) procInviteTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procInviteTopic")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_INVITE_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_INVITE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_INVITE_TOPIC_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	// resp_invite_topic false Join topic failure.

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_INVITE_TOPIC_CMD)
	resp.Time = time.Now().Unix()
	resp.Repo = cmd.GetReport()

	//群组ID
	topicId := cmd.GetArgs()[0]
	clientId := session.State.(*base.SessionState).ClientID
	friendList := []string{}
	for i := 1; i < len(cmd.GetArgs()); i++ {
		friendList = append(friendList, cmd.GetArgs()[i])
	}

	//判断群组是否存在
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result == nil {
		log.Error(info.TOPIC_DOES_NOT_EXISTS)
		self.respCmd(protocol.RESP_INVITE_TOPIC_CMD, session, cmd.GetReport(), false, info.TOPIC_DOES_NOT_EXISTS)
		return err
	}

	//判断用户是否属于该群组
	users := result.ClientsID
	if !common.InArray(users, clientId) {
		log.Error(info.YOU_WERE_NOT_IN_TOPIC)
		self.respCmd(protocol.RESP_INVITE_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_WERE_NOT_IN_TOPIC)
		return err
	}

	//加入
	readyToJoinTopic := []string{}
	for i := 0; i < len(friendList); i++ {
		if !common.InArray(users, friendList[i]) {
			users = append(users, friendList[i])
			readyToJoinTopic = append(readyToJoinTopic, friendList[i])
		}
	}

	//执行加入操作
	TopicStoreData := mongo_store.TopicStoreData{topicId, clientId, users}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(protocol.RESP_JOIN_TOPIC_CMD, session, cmd.GetReport(), false, info.JOIN_TOPIC_FAILURE)
		return err
	}

	if len(readyToJoinTopic) > 0 {
		for i := 0; i < len(readyToJoinTopic); i++ {
			resp.AddArg(readyToJoinTopic[i])
		}
	} else {
		resp.Ok = false
		resp.Message = info.JOIN_TOPIC_FAILURE
	}

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if len(readyToJoinTopic) > 0 {
		for i := 0; i < len(readyToJoinTopic); i++ {
			// user string, timeNow int64
			err = self.msgServer.mongoStore.MarkTopicRecordMessageFromUserAndTime(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, readyToJoinTopic[i], time.Now().Unix(), topicId)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	}

	return err
}

//离开Topic
func (self *ProtoProc) procLeaveTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procLeaveTopic")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_LEAVE_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_LEAVE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_LEAVE_TOPIC_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]
	clientId := session.State.(*base.SessionState).ClientID

	//如果群组存在,群组存在
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result == nil {
		log.Error(info.TOPIC_DOES_NOT_EXISTS)
		self.respCmd(protocol.RESP_LEAVE_TOPIC_CMD, session, cmd.GetReport(), false, info.TOPIC_DOES_NOT_EXISTS)
		return err
	}
	users := result.ClientsID
	if !common.InArray(users, clientId) {
		log.Error(info.YOU_WERE_NOT_IN_TOPIC)
		self.respCmd(protocol.RESP_LEAVE_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_WERE_NOT_IN_TOPIC)
		return err
	}
	users = common.DeleteChild(users, clientId)
	TopicStoreData := mongo_store.TopicStoreData{topicId, clientId, users}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(protocol.RESP_LEAVE_TOPIC_CMD, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	self.respCmd(protocol.RESP_LEAVE_TOPIC_CMD, session, cmd.GetReport(), true, "")
	return nil
}

//获取Topic信息
func (self *ProtoProc) procListTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procListTopic")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_LIST_TOPIC_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	clientId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_LIST_TOPIC_CMD)
	resp.Time = time.Now().Unix()
	resp.Repo = cmd.GetReport()

	//如果群组存在返回成员信息
	result := self.msgServer.mongoStore.GetTopicsFromClientID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, clientId)
	if result == nil {
		log.Info("No topic")
		self.respCmd(protocol.RESP_LIST_TOPIC_CMD, session, cmd.GetReport(), false, info.TOPIC_DOES_NOT_EXISTS)
		return err
	}
	topicsNameArr := make([]string, 0)
	for _, v := range result {
		topicsNameArr = append(topicsNameArr, v.TopicID)
	}

	temp, err := json.Marshal(topicsNameArr)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(protocol.RESP_LIST_TOPIC_CMD, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	resp.AddArg(string(temp))
	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

//获取Topic成员信息
func (self *ProtoProc) procTopicMembersList(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procTopicMembersList")
	var err error

	if session.State == nil {
		self.respCmd(protocol.RESP_TOPIC_MEMBERS_LIST_CMD, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_TOPIC_MEMBERS_LIST_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_TOPIC_MEMBERS_LIST_CMD, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]
	clientId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_TOPIC_MEMBERS_LIST_CMD)
	resp.Time = time.Now().Unix()
	resp.Repo = cmd.GetReport()

	//如果群组存在返回成员信息
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result == nil {
		log.Error(info.NO_CLIENTS_IN_TOPIC)
		self.respCmd(protocol.RESP_TOPIC_MEMBERS_LIST_CMD, session, cmd.GetReport(), false, info.NO_CLIENTS_IN_TOPIC)
		return err
	}

	//检查用户是否在群组里
	users := result.ClientsID
	if !common.InArray(users, clientId) {
		log.Error(info.YOU_WERE_NOT_IN_TOPIC)
		self.respCmd(protocol.RESP_TOPIC_MEMBERS_LIST_CMD, session, cmd.GetReport(), false, info.YOU_WERE_NOT_IN_TOPIC)
		return err
	}

	//获取群组成员信息
	clientInfo := self.msgServer.mongoStore.GetFriendsFromIds(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, users)
	if clientInfo == nil {
		log.Error("no client list")
		self.respCmd(protocol.RESP_TOPIC_MEMBERS_LIST_CMD, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	temp, err := json.Marshal(clientInfo)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(protocol.RESP_TOPIC_MEMBERS_LIST_CMD, session, cmd.GetReport(), false, info.ERROR)
		return err
	}
	resp.AddArg(topicId)
	resp.AddArg(string(temp))

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//处理Topic信息
func (self *ProtoProc) procSendMessageTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMessageTopic")
	var err error

	msgType := cmd.GetCmdName()

	if session.State == nil {
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_MESSAGE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	send2Msg := cmd.GetArgs()[0]
	topicId := cmd.GetArgs()[1]

	fromID := session.State.(*base.SessionState).ClientID
	send2Time := time.Now().Unix()

	uuid := common.NewV4().String()

	//获取Topic的信息
	topicResult := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if topicResult == nil {
		log.Error(info.TOPIC_DOES_NOT_EXISTS)
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.TOPIC_DOES_NOT_EXISTS)
		return err
	}

	//判断用户是否属于该Topic
	if !common.InArray(topicResult.ClientsID, fromID) {
		log.Info(fromID + " don't belong to the " + topicId)
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.YOU_WERE_NOT_IN_TOPIC)
		return err
	}

	//获取群组成员信息
	msgResult := self.msgServer.mongoStore.GetOnlineClientsFromIds(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, topicResult.ClientsID)
	if msgResult == nil {
		log.Info("no client list")
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	var clientGroup map[string][]mongo_store.SessionStoreData
	clientGroup = make(map[string][]mongo_store.SessionStoreData)
	onlineUsers := make([]string, 0) //所有在线用户

	for _, v := range msgResult {
		if clientGroup[v.MsgServerAddr] != nil {
			clientGroup[v.MsgServerAddr] = append(clientGroup[v.MsgServerAddr], *v)
		} else {
			clientGroup[v.MsgServerAddr] = make([]mongo_store.SessionStoreData, 1)
			clientGroup[v.MsgServerAddr][0] = *v
		}
		onlineUsers = append(onlineUsers, v.ClientID)
	}

	//保存消息到mongodb中
	data := mongo_store.TopicRecordMessageData{msgType, fromID, topicId, send2Msg, send2Time, uuid, []string{}}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, &data)
	if err != nil {
		log.Error(err.Error())
		self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
		return err
	}

	//直接到客户端的信息
	receive := protocol.NewCmdResponse(NCommendMappedMap[msgType].ReceiveCmd)
	receive.AddArg(send2Msg)
	receive.AddArg(topicId)
	receive.AddArg(fromID)
	receive.AddArg(strconv.FormatInt(send2Time, 10))
	receive.AddArg(uuid)

	// map[127.0.0.1:19001:[{aa 192.168.60.101:57826 127.0.0.1:19001 false} {bb 192.168.60.101:57829 127.0.0.1:19001 true} {cc 192.168.60.101:57845 127.0.0.1:19001 true}]]
	//&{SEND_MESSAGE_TOPIC [hello topics aa]}
	for k, v := range clientGroup {
		if k == self.msgServer.cfg.LocalIP {
			log.Info("In the same server")
			for _, client := range v {
				if self.msgServer.sessions[client.ClientID] != nil {
					err = self.msgServer.sessions[client.ClientID].AsyncSend(receive)
					if err != nil {
						log.Error(err.Error())
					}

					//缓存uuid,等待ack
					ack := new(base.AckFrequency)
					ack.Frequency = 1
					ack.LastTime = send2Time
					//用uuid+ClientID做key
					self.msgServer.topicAckMap[client.ClientID+uuid] = ack
				}
			}
		} else {
			log.Info("Not in the same server")
			sjm, err := json.Marshal(v)
			if err != nil {
				log.Error(err.Error())
				self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
				return err
			}

			//router转发的信息
			tempCmd := protocol.NewCmdSimple(NCommendMappedMap[msgType].RouterCmd)
			tempCmd.AddArg(send2Msg)
			tempCmd.AddArg(topicId)
			tempCmd.AddArg(fromID)
			//strconv.FormatInt(send2Time, 10)
			tempCmd.AddArg(strconv.FormatInt(send2Time, 10))
			tempCmd.AddArg(string(sjm))
			tempCmd.AddArg(uuid)

			jcmd, err := json.Marshal(tempCmd)
			if err != nil {
				log.Error(err.Error())
				self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), false, info.ERROR)
				return err
			}

			routerMsg := protocol.NewCmdSimple(protocol.ROUTE_MSG_CMD)
			routerMsg.AddArg(v[0].MsgServerAddr)
			routerMsg.AddArg(string(jcmd))

			if self.msgServer.channels[protocol.SYSCTRL_SEND] != nil {
				err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(routerMsg)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}

	self.respCmd(NCommendMappedMap[msgType].RespCmd, session, cmd.GetReport(), true, "")
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
		resp := protocol.NewCmdResponse(NCommendMappedMap[v.MsgType].ReceiveCmd)
		resp.AddArg(v.Content)
		resp.AddArg(v.ToID)
		resp.AddArg(v.FromID)
		resp.AddArg(strconv.FormatInt(v.Time, 10))
		resp.AddArg(v.UUID)

		time.Sleep(100)

		//缓存uuid,等待ack
		ack := new(base.AckFrequency)
		ack.Frequency = 1
		ack.LastTime = time.Now().Unix()
		self.msgServer.topicAckMap[cid+v.UUID] = ack

		if self.msgServer.sessions[cid] != nil {
			err = self.msgServer.sessions[cid].Send(resp)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	}

	return err
}

//Topic ACK
func (self *ProtoProc) procTopicAck(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procTopicAck")
	var err error

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.TOPIC_ACK_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	clientID := session.State.(*base.SessionState).ClientID
	uuid := cmd.GetArgs()[0]

	if self.msgServer.topicAckMap[clientID+uuid] != nil {
		msg := self.msgServer.mongoStore.ReadTopicRecordMessageFromUuid(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, uuid)
		if msg == nil {
			log.Info("No message")
			return err
		}

		//标记已读
		err = self.msgServer.mongoStore.MarkTopicRecordMessageFromUuid(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, uuid, append(msg.IsRead, clientID))
		if err != nil {
			log.Error(err.Error())
			return err
		}

		self.deleteWithMutex(self.msgServer.topicAckMap, clientID+uuid, self.msgServer.topicAckMutex)
	}

	return err
}

//TopicACK超时重发处理
func (self *ProtoProc) procTopicTimeoutRetransmission() {
	// log.Info("procTopicTimeoutRetransmission")

	for k, v := range self.msgServer.topicAckMap {
		//通过k截取到ClientID, 36为UUID的固定长度
		clientID := k[0 : len(k)-36]
		UUID := k[len(k)-36:]

		if v.Frequency >= protocol.TOPIC_ACK_FAILURES {
			log.Info(k + " is dead.")
			self.deleteWithMutex(self.msgServer.topicAckMap, k, self.msgServer.topicAckMutex)
			continue
		}
		if (time.Now().Unix() - v.LastTime) > protocol.TOPIC_ACK_TIMEOUT {
			//重设Ack
			self.msgServer.topicAckMap[k].Frequency++
			self.msgServer.topicAckMap[k].LastTime = time.Now().Unix()

			//从mongo读取信息
			recordData := self.msgServer.mongoStore.ReadTopicRecordMessageFromUuid(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, UUID)
			if recordData == nil {
				log.Info("No data.")
				continue
			}

			//重发信息
			resp := protocol.NewCmdResponse(NCommendMappedMap[recordData.MsgType].ReceiveCmd)
			resp.AddArg(recordData.Content)
			resp.AddArg(recordData.ToID)
			resp.AddArg(recordData.FromID)
			resp.AddArg(strconv.FormatInt(recordData.Time, 10))
			resp.AddArg(recordData.UUID)

			if self.msgServer.sessions[clientID] != nil {
				err := self.msgServer.sessions[clientID].Send(resp)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}
}

//用户退出关闭通道
func (self *ProtoProc) clientQuit(session *libnet.Session) {
	if session.State != nil {
		ClientID := session.State.(*base.SessionState).ClientID
		self.deleteWithMutex(self.msgServer.sessions, ClientID, self.msgServer.scanSessionMutex)
	}

	session.Close()
}

//删除锁
func (self *ProtoProc) deleteWithMutex(parent interface{}, children string, dMutex sync.Mutex) {
	dMutex.Lock()
	defer dMutex.Unlock()

	switch parent.(type) {
	case base.SessionMap:
		delete(parent.(base.SessionMap), children)
	case base.AckMap:
		delete(parent.(base.AckMap), children)
	}
}

//回复客户端统一
func (self *ProtoProc) respCmd(respCmd string, session *libnet.Session, repo interface{}, ok bool, message string) {
	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(respCmd)
	resp.Repo = repo
	resp.Ok = ok
	resp.Message = message
	resp.Time = time.Now().Unix()

	//返回用户请求
	err := session.Send(resp)
	if err != nil {
		log.Error("report :", err.Error())
	}
}
