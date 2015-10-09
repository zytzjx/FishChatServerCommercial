package main

import (
	"flag"
	"goProject/base"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	"sync"
	"time"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type MsgServer struct {
	cfg      *MsgServerConfig
	sessions base.SessionMap
	channels base.ChannelMap
	// topics   protocol.TopicMap
	server *libnet.Server

	p2pAckMap        base.AckMap
	topicAckMap      base.AckMap
	mutualAckMap     base.AckMap
	scanSessionMutex sync.Mutex
	p2pAckMutex      sync.Mutex
	topicAckMutex    sync.Mutex
	mutualAckMutex   sync.Mutex

	mongoStore *mongo_store.MongoStore
}

func NewMsgServer(cfg *MsgServerConfig) *MsgServer {
	return &MsgServer{
		cfg:      cfg,
		sessions: make(base.SessionMap),
		channels: make(base.ChannelMap),
		// topics:       make(protocol.TopicMap),
		server:       new(libnet.Server),
		p2pAckMap:    make(base.AckMap),
		topicAckMap:  make(base.AckMap),
		mutualAckMap: make(base.AckMap),
		mongoStore:   mongo_store.NewMongoStore(cfg.Mongo.Addr, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password),
	}
}

//创建Channels
func (self *MsgServer) createChannels() {
	log.Info("createChannels")
	for _, c := range base.ChannleList {
		channel := libnet.NewChannel()
		self.channels[c] = base.NewChannelState(c, channel)
	}
}

//统计数据
func (self *MsgServer) sendMonitorData() error {
	log.Info("sendMonitorData")

	mb := NewMonitorBeat("monitor", self.cfg.MonitorBeatTime, 40, 10)

	if self.channels[protocol.SYSCTRL_MONITOR] != nil {
		for {
			resp := protocol.NewCmdMonitor(protocol.TYPE_MSG_SERVER_SERVER_CMD, time.Now().Unix(),
				protocol.MsgServerMonitorData{
					SessionNum: (uint64)(len(self.sessions)),
				})
			mb.Beat(self.channels[protocol.SYSCTRL_MONITOR].Channel, resp)
		}
	}

	return nil
}

// 扫描进程
func (self *MsgServer) scanDeadSession() {
	log.Info("scanDeadSession")
	timer := time.NewTicker(self.cfg.ScanDeadSessionTimeout * time.Second)
	ttl := time.After(self.cfg.Expire * time.Second)
	for {
		select {
		case <-timer.C:
			// log.Info("scanDeadSession timeout")
			go func() {
				for id, s := range self.sessions {

					if (s.State).(*base.SessionState).Alive == false {
						log.Info("delete" + id)

						self.scanSessionMutex.Lock()
						delete(self.sessions, id)
						self.scanSessionMutex.Unlock()
					} else {
						s.State.(*base.SessionState).Alive = false
					}

					self.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, id, false)
				}
			}()
		case <-ttl:
			break
		}
	}
}

//扫描超时仍未返回的ack,重发消息
func (self *MsgServer) scanTimeoutAck() {
	log.Info("scanTimeoutAck")
	timer := time.NewTicker(self.cfg.ScanTimeoutAck * time.Second)
	ttl := time.After(self.cfg.Expire * time.Second)
	pp := NewProtoProc(self)
	for {
		select {
		case <-timer.C:
			// log.Info("scanTimeoutAck timeout")
			go func() {
				//P2P信息超时重发
				pp.procP2pTimeoutRetransmission()
				//Topic信息超时重发
				pp.procTopicTimeoutRetransmission()
			}()
		case <-ttl:
			break
		}
	}
}

//协议解析
func (self *MsgServer) parseProtocol(cmd protocol.CmdSimple, session *libnet.Session) error {
	var err error
	pp := NewProtoProc(self)

	cmdName := cmd.GetCmdName()

	switch cmdName {
	//PING
	case protocol.SEND_PING_CMD:
		err = pp.procPing(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//router订阅
	case protocol.SUBSCRIBE_CHANNEL_CMD:
		pp.procSubscribeChannel(&cmd, session)

	//router过来的信息统一接收端口
	case protocol.ROUTE_MSG_CMD:
		err = pp.procRouteMsg(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//登陆
	case protocol.SEND_CLIENT_ID_CMD:
		err = pp.procClientID(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//退出
	case protocol.SEND_LOGOUT_CMD:
		err = pp.procLogout(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//P2P信息
	case protocol.SEND_MESSAGE_P2P_CMD:
		err = pp.procSendMessageP2P(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//创建Topic
	case protocol.SEND_CREATE_TOPIC_CMD:
		err = pp.procCreateTopic(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//加入Topic
	case protocol.SEND_JOIN_TOPIC_CMD:
		err = pp.procJoinTopic(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//离开Topic
	case protocol.SEND_LEAVE_TOPIC_CMD:
		err = pp.procLeaveTopic(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//获取群组信息
	case protocol.SEND_LIST_TOPIC_CMD:
		err = pp.procListTopic(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	//获取Topic成员
	case protocol.SEND_TOPIC_MEMBERS_LIST_CMD:
		err = pp.procTopicMembersList(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//Topic信息处理
	case protocol.SEND_MESSAGE_TOPIC_CMD:
		err = pp.procSendMessageTopic(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	// p2p ack
	case protocol.P2P_ACK_CMD:
		err = pp.procP2pAck(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	// topic ack
	case protocol.TOPIC_ACK_CMD:
		err = pp.procTopicAck(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//view friend list
	case protocol.SEND_VIEW_FRIENDS_CMD:
		err = pp.procViewFriends(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//add friend
	case protocol.SEND_ADD_FRIEND_CMD:
		err = pp.procAddFriend(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//del friend
	case protocol.SEND_DEL_FRIEND_CMD:
		err = pp.procDelFriend(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//请求统一接口
	case protocol.SEND_ASK_CMD:
		err = pp.procAsk(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	//接收请求统一接口
	case protocol.SEND_REACT_CMD:
		err = pp.procReact(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}

	default:
		log.Info(cmd.GetCmdName())
	}

	return err
}
