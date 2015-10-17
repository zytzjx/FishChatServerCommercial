package main

import (
	"encoding/json"
	"errors"
	"goProject/base"
	"goProject/common"
	"goProject/info"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	// "gopkg.in/mgo.v2/bson"
)

// 查询好友
func (self *ProtoProc) procViewFriends(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procViewFriends")
	var err error

	if session.State == nil {
		return nil
	}
	clientId := session.State.(*base.SessionState).ClientID

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
	if len(cmd.GetArgs()) < protocol.SEND_ADD_FRIEND_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_ADD_FRIEND_CMD)
	resp.Repo = cmd.GetReport()

	clientId := session.State.(*base.SessionState).ClientID
	friendId := cmd.GetArgs()[0]

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
	if len(cmd.GetArgs()) < protocol.SEND_DEL_FRIEND_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	//定义返回用户请求信息
	resp := protocol.NewCmdResponse(protocol.RESP_DEL_FRIEND_CMD)
	resp.Repo = cmd.GetReport()

	clientId := session.State.(*base.SessionState).ClientID
	friendId := cmd.GetArgs()[0]

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
