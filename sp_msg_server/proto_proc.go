package main

import (
	"encoding/json"
	"errors"
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

//接收用户登录ID
func (self *ProtoProc) procClientID(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procClientID")
	var err error

	//返回信息
	resp := protocol.NewCmdResponse(protocol.RESP_CLIENT_ID_CMD)
	resp.Repo = cmd.GetReport()

	if len(cmd.GetArgs()) < 1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.clientQuit(session)
		return nil
	}

	ClientID := cmd.GetArgs()[0]
	if ClientID != "88888888" {
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

	if len(cmd.GetArgs()) < protocol.SEND_MESSAGE_P2P_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	fromID := cmd.GetArgs()[0]
	send2Msg := cmd.GetArgs()[1]
	send2ID := cmd.GetArgs()[2]
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

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}

// 增加Topic
func (self *ProtoProc) procCreateTopic(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procCreateTopic")
	var err error

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_CREATE_TOPIC_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	founderId := cmd.GetArgs()[0]
	//群组ID
	topicId := cmd.GetArgs()[1]

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
	if len(cmd.GetArgs()) < protocol.SEND_JOIN_TOPIC_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	clientId := cmd.GetArgs()[0]
	//群组ID
	topicId := cmd.GetArgs()[1]

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
	if len(cmd.GetArgs()) < protocol.SEND_LEAVE_TOPIC_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	clientId := cmd.GetArgs()[0]
	//群组ID
	topicId := cmd.GetArgs()[1]

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

	if len(cmd.GetArgs()) < 1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	clientId := cmd.GetArgs()[0]

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
	if len(cmd.GetArgs()) < protocol.SEND_TOPIC_MEMBERS_LIST_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	clientId := cmd.GetArgs()[0]
	//群组ID
	topicId := cmd.GetArgs()[1]

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

// 查询好友
func (self *ProtoProc) procViewFriends(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procViewFriends")
	var err error

	if len(cmd.GetArgs()) < 1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	clientId := cmd.GetArgs()[0]

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_VIEW_FRIENDS_CMD)
	resp.Repo = cmd.GetReport()

	//查询用户好友ID
	clientInfo, err := self.msgServer.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientId)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	//查询好友列表信息
	result := self.msgServer.mongoStore.GetFriendsFromIds(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientInfo.Friends)
	if result == nil {
		log.Info("no client list")
		result = []*mongo_store.SessionStoreDataFriends{}
	}

	temp, err := json.Marshal(result)
	if err != nil {
		log.Error(err.Error())
		resp.Message = info.YOU_HAVE_NO_FRIENDS
		resp.Ok = true
	} else {
		resp.AddArg(string(temp))
		if self.msgServer.sessions[clientId] != nil {
			err = self.msgServer.sessions[clientId].Send(resp)
			if err != nil {
				log.Error(err.Error())
				return err
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

//添加好友
func (self *ProtoProc) procAddFriend(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procAddFriend")
	var err error
	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_ADD_FRIEND_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_ADD_FRIEND_CMD)
	resp.Repo = cmd.GetReport()

	clientId := cmd.GetArgs()[0]
	friendId := cmd.GetArgs()[1]

	clientInfo, err := self.msgServer.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientId)
	if err != nil {
		log.Error(err.Error())
		resp.Message = info.NO_CLIENT_INFO
		resp.Ok = false
	} else {
		friends := clientInfo.Friends
		if common.InArray(friends, friendId) {
			resp.Message = info.THE_ID_IS_ALREADY_YOUR_FRIEND
			resp.Ok = false
		} else {
			_, err := self.msgServer.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, friendId)
			if err != nil {
				notFound := errors.New("not found")
				if err.Error() == notFound.Error() {
					log.Error(err.Error())
					resp.Message = info.THIS_ID_IS_NOT_EXISTS
					resp.Ok = false
				} else {
					log.Error(err.Error())
					resp.Message = info.ERROR
					resp.Ok = false
				}
			} else {
				err = self.msgServer.mongoStore.UpdateFriendsFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientId, append(friends, friendId))
				if err != nil {
					log.Error(err.Error())
					resp.Message = info.ERROR
					resp.Ok = false
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

// 删除好友
func (self *ProtoProc) procDelFriend(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procDelFriend")
	var err error

	if session.State == nil {
		return nil
	}
	if len(cmd.GetArgs()) < protocol.SEND_DEL_FRIEND_CMD_ARGS_NUM+1 {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_DEL_FRIEND_CMD)
	resp.Repo = cmd.GetReport()

	clientId := cmd.GetArgs()[0]
	friendId := cmd.GetArgs()[1]

	clientInfo, err := self.msgServer.mongoStore.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientId)
	if err != nil {
		log.Error(err.Error())
		resp.Message = info.NO_CLIENT_INFO
		resp.Ok = false
	} else {
		friends := clientInfo.Friends
		if !common.InArray(friends, friendId) {
			resp.Message = info.YOU_HAVE_NO_THIS_FRIEND
			resp.Ok = false
		} else {
			friends = common.DeleteChild(friends, friendId)
			err = self.msgServer.mongoStore.UpdateFriendsFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientId, friends)
			if err != nil {
				log.Error(err.Error())
				resp.Message = info.ERROR
				resp.Ok = false
			}
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

//用户退出关闭通道
func (self *ProtoProc) clientQuit(session *libnet.Session) {
	if session.State != nil {
		ClientID := session.State.(*base.SessionState).ClientID
		delete(self.msgServer.sessions, ClientID)
	}

	session.Close()
}
