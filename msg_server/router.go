package main

import (
	"encoding/json"
	"goProject/base"
	// "goProject/common"
	"goProject/info"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	// "gopkg.in/mgo.v2/bson"
	// "strconv"
	"time"
)

func (self *ProtoProc) procRouteMsg(cmd protocol.Cmd, session *libnet.Session) error {
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_MSG_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	bmsg := cmd.GetArgs()[1]
	msg := new(protocol.CmdSimple)
	err = json.Unmarshal([]byte(bmsg), &msg)
	if err != nil {
		log.Error("error:", err)
		return err
	}

	cmdName := msg.GetCmdName()

	switch cmdName {
	//router踢用户下线
	case protocol.ROUTE_CHANGE_MESSAGE_SERVER_CMD:
		err = self.procRouterChangeMsgServer(msg, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//路由P2P信息
	case protocol.ROUTE_MESSAGE_P2P_CMD:
		err = self.procRouteMessageP2P(msg, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	case protocol.ROUTE_NOTIFY_P2P_CMD:
		err = self.procRouteMessageP2P(msg, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//Topic route信息
	case protocol.ROUTE_MESSAGE_TOPIC_CMD:
		err = self.procRouteMessageTopic(msg, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	case protocol.ROUTE_NOTIFY_TOPIC_CMD:
		err = self.procRouteMessageTopic(msg, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//ROUTE 请求统一接口
	case protocol.ROUTE_ASK_CMD:
		err = self.procRouteAsk(msg, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	default:
		log.Info(msg)
	}

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

//解析Router P2P信息
func (self *ProtoProc) procRouteMessageP2P(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteMessageP2P")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_MESSAGE_P2P_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	msgType := cmd.GetCmdName()
	// &{route_message_p2p [faqowetgmokawngh aa bb 2015-08-18 09:28:51 -0400 EDT 7f152006-6909-4955-bdb4-92f0e3cb354e]
	send2Msg := cmd.GetArgs()[0]
	fromID := cmd.GetArgs()[1]
	send2ID := cmd.GetArgs()[2]
	send2Time := cmd.GetArgs()[3]
	uuid := cmd.GetArgs()[4]

	log.Info(msgType)

	// &{ROUTE_MESSAGE_P2P [awge4y aa bb 锟?9170c3a9-40c4-420c-80c8-756c19efde16]}
	receive := protocol.NewCmdResponse(NCommendMappedMap[msgType].ReceiveCmd)
	receive.AddArg(send2Msg)
	receive.AddArg(fromID)
	receive.AddArg(send2Time)
	receive.AddArg(uuid)

	if self.msgServer.sessions[send2ID] != nil {
		self.msgServer.sessions[send2ID].Send(receive)
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

//route
func (self *ProtoProc) procRouteAsk(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procRouteAsk")
	var err error

	if len(cmd.GetArgs()) < protocol.ROUTE_ASK_CMD_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	msgType := cmd.GetArgs()[0]
	fromID := cmd.GetArgs()[1]
	toID := cmd.GetArgs()[2]
	msgtime := cmd.GetArgs()[3]
	uuid := cmd.GetArgs()[4]

	receive := protocol.NewCmdResponse(NCommendMappedMap[msgType].ReceiveCmd)
	receive.AddArg(msgType)
	receive.AddArg(fromID)
	receive.AddArg(msgtime)
	receive.AddArg(uuid)

	if self.msgServer.sessions[toID] != nil {
		self.msgServer.sessions[toID].Send(receive)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		//储存ACK，用来验证
		ack := new(base.AckFrequency)
		ack.Frequency = 1
		ack.LastTime = time.Now().Unix()
		self.msgServer.mutualAckMap[uuid] = ack
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
	msgType := cmd.GetCmdName()
	send2Msg, topicId, fromID, send2Time, getTargetListStr, uuid := args[0], args[1], args[2], args[3], args[4], args[5]

	getTargetListByte := []byte(getTargetListStr)
	var Clients []mongo_store.SessionStoreData
	json.Unmarshal(getTargetListByte, &Clients)

	for _, v := range Clients {
		//router转发的信息
		newCmd := protocol.NewCmdResponse(NCommendMappedMap[msgType].ReceiveCmd)
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
			err = self.msgServer.sessions[v.ClientID].AsyncSend(newCmd)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	return nil
}
