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
	"strconv"
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

	//返回信息
	resp := protocol.NewCmdResponse(protocol.RESP_PONG_CMD)
	resp.Repo = cmd.GetReport()

	if session.State == nil {
		//PONG
		err := session.Send(resp)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return nil
	}

	cid := session.State.(*base.SessionState).ClientID
	self.msgServer.scanSessionMutex.Lock()
	defer self.msgServer.scanSessionMutex.Unlock()

	if self.msgServer.sessions[cid] == nil || self.msgServer.sessions[cid].State == nil {
		return nil
	}

	self.msgServer.sessions[cid].State.(*base.SessionState).Alive = true

	self.msgServer.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, cid, true)

	//PONG
	err := session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

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
		receive := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_P2P_CMD)
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

	//返回信息
	resp := protocol.NewCmdResponse(protocol.RESP_CLIENT_ID_CMD)
	resp.Repo = cmd.GetReport()

	if len(cmd.GetArgs()) < protocol.SEND_CLIENT_ID_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.clientQuit(session)
		return nil
	}
	ClientID := cmd.GetArgs()[0]
	if len(ClientID) < 1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
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
				routerMsg := protocol.NewCmdResponse(protocol.SEND_CHANGE_MESSAGE_SERVER_CMD)
				routerMsg.AddArg(ClientID)
				routerMsg.AddArg(clientInfo.MsgServerAddr)
				err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(routerMsg)
				if err != nil {
					log.Error(err.Error())
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
					delete(self.msgServer.sessions, ClientID)
				}
			}
		}

		sessionStoreData := mongo_store.SessionStoreData{ClientID, session.Conn().RemoteAddr().String(),
			self.msgServer.cfg.LocalIP, clientInfo.Friends, true}

		err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
		if err != nil {
			log.Error(err.Error())
			resp.Message = info.ERROR
			resp.Ok = false
		}
	} else {
		sessionStoreData := mongo_store.SessionStoreData{ClientID, session.Conn().RemoteAddr().String(),
			self.msgServer.cfg.LocalIP, []string{}, true}

		// update login info
		err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
		if err != nil {
			log.Error(err.Error())
			resp.Message = info.ERROR
			resp.Ok = false
		}
	}

	self.msgServer.sessions[ClientID] = session
	self.msgServer.sessions[ClientID].State = base.NewSessionState(ClientID, true,
		session.Conn().RemoteAddr().String(), self.msgServer.cfg.LocalIP)

	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
	}

	//获取用户未读信息
	self.procOfflineMsg(session, ClientID)

	return err
}

//更换MsgServer服务器
func (self *ProtoProc) procRouterChangeMsgServer(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouterChangeMsgServer")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_CHANGE_MESSAGE_SERVER_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return err
	}

	clientID := cmd.GetArgs()[0]

	if self.msgServer.sessions[clientID] == nil {
		log.Info("the user is not login in this msg server.")
		return nil
	} else {
		resp := protocol.NewCmdResponse(protocol.RESP_LOGOUT_CMD)
		resp.Message = info.YOU_HAVE_TO_RE_LOGIN

		err = self.msgServer.sessions[clientID].Send(resp)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		self.clientQuit(self.msgServer.sessions[clientID])
	}

	return err
}

//退出登录
func (self *ProtoProc) procLogout(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procLogout")
	var err error

	if session.State == nil {
		return nil
	}
	clientID := session.State.(*base.SessionState).ClientID
	if clientID != "" {
		// 标记用户离线
		err := self.msgServer.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientID, false)
		if err != nil {
			log.Error(err.Error())
		}
	}

	//返回用户退出成功信息
	resp := protocol.NewCmdResponse(protocol.RESP_LOGOUT_CMD)
	resp.Repo = cmd.GetReport()
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	self.clientQuit(session)
	return err
}

// 解析P2P信息
func (self *ProtoProc) procSendMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendMessageP2P")
	var err error

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_MESSAGE_P2P_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	fromID := session.State.(*base.SessionState).ClientID
	send2Msg := cmd.GetArgs()[0]
	send2ID := cmd.GetArgs()[1]
	send2Time := time.Now().Unix()
	uuid := common.NewV4().String()

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_MESSAGE_P2P_CMD)
	resp.Repo = cmd.GetReport()

	//保存消息到mongodb中
	data := mongo_store.P2PRecordMessageData{fromID, send2ID, send2Msg, send2Time, uuid, false}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, &data)
	if err != nil {
		log.Error(err.Error())
		resp.Message = info.ERROR
		resp.Ok = false
	} else {
		if self.msgServer.sessions[send2ID] != nil {
			log.Info("In the same server")

			receive := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_P2P_CMD)
			receive.AddArg(send2Msg)
			receive.AddArg(fromID)
			receive.AddArg(strconv.FormatInt(send2Time, 10))
			receive.AddArg(uuid)

			if self.msgServer.sessions[send2ID] != nil {
				self.msgServer.sessions[send2ID].Send(receive)
				if err != nil {
					log.Error(err.Error())
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

				rcmd := protocol.NewCmdResponse(protocol.SEND_MESSAGE_P2P_CMD)
				rcmd.AddArg(send2Msg)
				rcmd.AddArg(fromID)
				rcmd.AddArg(send2ID)
				rcmd.AddArg(strconv.FormatInt(send2Time, 10))
				rcmd.AddArg(uuid)

				err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(rcmd)
				if err != nil {
					log.Error(err.Error())
					return err
				}
			}
		}
	}

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}

//解析Router P2P信息
func (self *ProtoProc) procRouteMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMessageP2P")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_MESSAGE_P2P_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	// &{route_message_p2p [faqowetgmokawngh aa bb 2015-08-18 09:28:51 -0400 EDT 7f152006-6909-4955-bdb4-92f0e3cb354e]
	send2Msg := cmd.GetArgs()[0]
	fromID := cmd.GetArgs()[1]
	send2ID := cmd.GetArgs()[2]
	send2Time := cmd.GetArgs()[3]
	uuid := cmd.GetArgs()[4]

	// &{ROUTE_MESSAGE_P2P [awge4y aa bb 锟?9170c3a9-40c4-420c-80c8-756c19efde16]}
	resp := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_P2P_CMD)
	resp.AddArg(send2Msg)
	resp.AddArg(fromID)
	resp.AddArg(send2Time)
	resp.AddArg(uuid)

	if self.msgServer.sessions[send2ID] != nil {
		self.msgServer.sessions[send2ID].Send(resp)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		ack := new(base.AckFrequency)
		ack.Frequency = 1
		ack.LastTime = time.Now().Unix()
		self.msgServer.p2pAckMap[uuid] = ack
	}

	return nil
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
	self.msgServer.p2pAckMutex.Lock()
	defer self.msgServer.p2pAckMutex.Unlock()

	if self.msgServer.p2pAckMap[uuid] != nil {
		//InACK
		log.Info(uuid + " inACK list")
		//标记已读
		err = self.msgServer.mongoStore.MarkP2PRecordMessageFromUuid(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_P2P_MESSAGE_COLLECTION, uuid)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		delete(self.msgServer.p2pAckMap, uuid)
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
			delete(self.msgServer.p2pAckMap, k)
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

			receive := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_P2P_CMD)
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
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_CREATE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]

	founderId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_CREATE_TOPIC_CMD)
	resp.Repo = cmd.GetReport()

	// 如果群组不存在,才添加群组
	if result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId); result == nil {
		//要存入数据库的数据
		ClientsID := []string{founderId}
		TopicStoreData := mongo_store.TopicStoreData{topicId, self.msgServer.cfg.LocalIP, founderId, ClientsID, true}

		err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
		if err != nil {
			log.Error(err.Error())
			resp.Message = info.CREATE_TOPIC_FAILURE
			resp.Ok = false
		}
	} else {
		resp.Message = info.TOPIC_ALREADY_EXISTS
		resp.Ok = false
	}

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//加入Topic
func (self *ProtoProc) procJoinTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procJoinTopic")
	var err error

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_JOIN_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]

	clientId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_JOIN_TOPIC_CMD)
	resp.Repo = cmd.GetReport()

	//如果群组存在,群组存在
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result != nil {
		users := result.ClientsID
		if !common.InArray(users, clientId) {
			users = append(users, clientId)
			TopicStoreData := mongo_store.TopicStoreData{topicId, self.msgServer.cfg.LocalIP, clientId, users, true}
			err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
			if err != nil {
				log.Error(err.Error())
				resp.Message = info.JOIN_TOPIC_FAILURE
				resp.Ok = false
			}
		} else {
			resp.Message = info.JOIN_TOPIC_FAILURE
			resp.Ok = false
		}
	} else {
		resp.Message = info.TOPIC_DOES_NOT_EXISTS
		resp.Ok = false
	}

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//离开Topic
func (self *ProtoProc) procLeaveTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procLeaveTopic")
	var err error

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_LEAVE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]

	clientId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_LEAVE_TOPIC_CMD)
	resp.Repo = cmd.GetReport()

	//如果群组存在,群组存在
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if result != nil {
		users := result.ClientsID
		if common.InArray(users, clientId) {
			users = common.DeleteChild(users, clientId)
			TopicStoreData := mongo_store.TopicStoreData{topicId, self.msgServer.cfg.LocalIP, clientId, users, true}
			err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, &TopicStoreData)
			if err != nil {
				log.Error(err.Error())
				resp.Message = info.JOIN_TOPIC_FAILURE
				resp.Ok = false
			}
		} else {
			resp.Message = info.YOU_WERE_NOT_IN_TOPIC
			resp.Ok = false
		}
	} else {
		resp.Message = info.TOPIC_DOES_NOT_EXISTS
		resp.Ok = false
	}

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//获取Topic信息
func (self *ProtoProc) procListTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procListTopic")
	var err error

	if session.State == nil {
		return nil
	}
	clientId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_LIST_TOPIC_CMD)
	resp.Repo = cmd.GetReport()

	//如果群组存在返回成员信息
	result := self.msgServer.mongoStore.GetTopicsFromClientID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, clientId)
	if result == nil {
		log.Info("No topic")
	} else {
		topicsNameArr := make([]string, 0)
		for _, v := range result {
			topicsNameArr = append(topicsNameArr, v.TopicID)
		}

		temp, err := json.Marshal(topicsNameArr)
		if err != nil {
			log.Error(err.Error())
			resp.Message = info.ERROR
			resp.Ok = false
		} else {
			resp.AddArg(string(temp))
		}
	}
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
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_TOPIC_MEMBERS_LIST_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//群组ID
	topicId := cmd.GetArgs()[0]

	clientId := session.State.(*base.SessionState).ClientID

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_TOPIC_MEMBERS_LIST_CMD)
	resp.Repo = cmd.GetReport()

	//如果群组存在返回成员信息
	result := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)

	if result != nil {
		users := result.ClientsID
		if common.InArray(users, clientId) {
			//获取群组成员信息
			clientInfo := self.msgServer.mongoStore.GetFriendsFromIds(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, users)
			if clientInfo == nil {
				log.Info("no client list")
				resp.Message = info.NO_CLIENTS_IN_TOPIC
				resp.Ok = false
			} else {
				temp, err := json.Marshal(clientInfo)
				if err != nil {
					log.Error(err.Error())
					resp.Message = info.NO_CLIENTS_IN_TOPIC
					resp.Ok = false
				} else {
					resp.AddArg(topicId)
					resp.AddArg(string(temp))

					if self.msgServer.sessions[clientId] != nil {
						err = self.msgServer.sessions[clientId].Send(resp)
						if err != nil {
							log.Error(err.Error())
							return err
						}
					}
				}
			}
		} else {
			resp.Message = info.YOU_WERE_NOT_IN_TOPIC
			resp.Ok = false
		}
	}

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

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_MESSAGE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	send2Msg := cmd.GetArgs()[0]
	topicId := cmd.GetArgs()[1]

	fromID := session.State.(*base.SessionState).ClientID
	send2Time := time.Now().Unix()

	uuid := common.NewV4().String()

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_MESSAGE_P2P_CMD)
	resp.Repo = cmd.GetReport()

	//获取Topic的信息
	topicResult := self.msgServer.mongoStore.GetTopicFromTopicID(mongo_store.DATA_BASE_NAME, mongo_store.TOPIC_INFO_COLLECTION, topicId)
	if topicResult == nil {
		log.Info("no topic in db")

		//返回用户请求
		err = session.Send(resp)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		return err
	}

	//判断用户是否属于该Topic
	if !common.InArray(topicResult.ClientsID, fromID) {
		log.Info(fromID + " don't belong to the " + topicId)

		//返回用户请求
		err = session.Send(resp)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return err
	}

	//获取群组成员信息
	msgResult := self.msgServer.mongoStore.GetOnlineClientsFromIds(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, topicResult.ClientsID)
	if msgResult == nil {
		log.Info("no client list")
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
	data := mongo_store.TopicRecordMessageData{fromID, topicId, send2Msg, send2Time, uuid, []string{}}
	err = self.msgServer.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, &data)
	if err != nil {
		log.Error(err.Error())
		//返回用户请求
		err = session.Send(resp)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return err
	}
	log.Info(fromID + " send to " + topicId + " of the data has been saved to the database.")

	//直接到客户端的信息
	receive := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_TOPIC_CMD)
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
					err = self.msgServer.sessions[client.ClientID].Send(receive)
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
			temp, err := json.Marshal(v)
			if err != nil {
				log.Error(err.Error())
				//返回用户请求
				err = session.Send(resp)
				if err != nil {
					log.Error(err.Error())
					return err
				}
				return err
			}

			//router转发的信息
			routerCmd := protocol.NewCmdResponse(protocol.SEND_MESSAGE_TOPIC_CMD)
			routerCmd.AddArg(send2Msg)
			routerCmd.AddArg(topicId)
			routerCmd.AddArg(fromID)
			//strconv.FormatInt(send2Time, 10)
			routerCmd.AddArg(strconv.FormatInt(send2Time, 10))
			routerCmd.AddArg(string(temp))
			routerCmd.AddArg(uuid)

			if self.msgServer.channels[protocol.SYSCTRL_SEND] != nil {
				err = self.msgServer.channels[protocol.SYSCTRL_SEND].Channel.Broadcast(routerCmd)
				if err != nil {
					log.Error(err.Error())
				}
			}
		}
	}

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return err
}

//解析Router Topic信息
func (self *ProtoProc) procRouteMessageTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMessageTopic")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_MESSAGE_TOPIC_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	args := cmd.GetArgs()
	send2Msg, topicId, fromID, send2Time, getTargetListStr, uuid := args[0], args[1], args[2], args[3], args[4], args[5]

	getTargetListByte := []byte(getTargetListStr)
	var Clients []mongo_store.SessionStoreData
	json.Unmarshal(getTargetListByte, &Clients)

	for _, v := range Clients {
		//router转发的信息
		newCmd := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_TOPIC_CMD)
		newCmd.AddArg(send2Msg)
		newCmd.AddArg(topicId)
		newCmd.AddArg(fromID)
		newCmd.AddArg(send2Time)
		newCmd.AddArg(uuid)

		//缓存uuid,等待ack
		ack := new(base.AckFrequency)
		ack.Frequency = 1
		ack.LastTime = time.Now().Unix()
		// ack.ClientID = v.ClientID
		self.msgServer.topicAckMap[v.ClientID+uuid] = ack

		if self.msgServer.sessions[v.ClientID] != nil {
			err = self.msgServer.sessions[v.ClientID].Send(newCmd)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	return nil
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
		resp := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_TOPIC_CMD)
		resp.AddArg(v.Content)
		resp.AddArg(v.ToID)
		resp.AddArg(v.FromID)
		resp.AddArg(strconv.FormatInt(v.Time, 10))
		resp.AddArg(v.UUID)

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
	self.msgServer.topicAckMutex.Lock()
	defer self.msgServer.topicAckMutex.Unlock()

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

		delete(self.msgServer.topicAckMap, clientID+uuid)
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
			delete(self.msgServer.topicAckMap, k)
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
			resp := protocol.NewCmdResponse(protocol.RECEIVE_MESSAGE_TOPIC_CMD)
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
		delete(self.msgServer.sessions, ClientID)
	}

	session.Close()
}
