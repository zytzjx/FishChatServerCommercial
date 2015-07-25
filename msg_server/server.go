

package main

import (
	"time"
	"flag"
	"sync"
	"goProject/log"
	"goProject/libnet"
	"goProject/base"
	"goProject/protocol"
	"goProject/storage/mongo_store"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type MsgServer struct {
	cfg               *MsgServerConfig
	sessions          base.SessionMap
	channels          base.ChannelMap
	topics            protocol.TopicMap
	server            *libnet.Server

	p2pAckStatus      base.AckMap
	scanSessionMutex  sync.Mutex
	p2pAckMutex       sync.Mutex
	
	mongoStore        *mongo_store.MongoStore
}

func NewMsgServer(cfg *MsgServerConfig) *MsgServer {
	return &MsgServer {
		cfg                : cfg,
		sessions           : make(base.SessionMap),
		channels           : make(base.ChannelMap),
		topics             : make(protocol.TopicMap),
		server             : new(libnet.Server),
		p2pAckStatus       : make(base.AckMap),
		mongoStore         : mongo_store.NewMongoStore(cfg.Mongo.Addr, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password),
	}
}

func (self *MsgServer)createChannels() {
	log.Info("createChannels")
	for _, c := range base.ChannleList {
		channel := libnet.NewChannel()
		self.channels[c] = base.NewChannelState(c, channel)
	}
}

func (self *MsgServer)sendMonitorData() error {
	log.Info("sendMonitorData")
	resp := protocol.NewCmdMonitor()

	mb := NewMonitorBeat("monitor", self.cfg.MonitorBeatTime, 40, 10)
	
	if self.channels[protocol.SYSCTRL_MONITOR] != nil {
		for{
			resp.SessionNum = (uint64)(len(self.sessions))
			mb.Beat(self.channels[protocol.SYSCTRL_MONITOR].Channel, resp)
		} 
	}

	return nil
}

func (self *MsgServer)scanDeadSession() {
	log.Info("scanDeadSession")
	timer := time.NewTicker(self.cfg.ScanDeadSessionTimeout * time.Second)
	ttl := time.After(self.cfg.Expire * time.Second)
	for {
		select {
		case <-timer.C:
			log.Info("scanDeadSession timeout")
			go func() {
				for id, s := range self.sessions {
					self.scanSessionMutex.Lock()
					//defer self.scanSessionMutex.Unlock()
					if (s.State).(*base.SessionState).Alive == false {
						log.Info("delete" + id)
						delete(self.sessions, id)
						self.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, id, false)
					} else {
						s.State.(*base.SessionState).Alive = false
						self.mongoStore.UpdateSessionAlive(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, id, false)
					}
					self.scanSessionMutex.Unlock()
				}
			}()
		case <-ttl:
			break
		}
	}
}

func (self *MsgServer)parseProtocol(cmd protocol.CmdSimple, session *libnet.Session) error {	
	var err error
	pp := NewProtoProc(self)

	switch cmd.GetCmdName() {
		case protocol.SEND_PING_CMD:
			err = pp.procPing(&cmd, session)
			if err != nil {
				log.Error("error:", err)
				return err
			}
		case protocol.SUBSCRIBE_CHANNEL_CMD:
			pp.procSubscribeChannel(&cmd, session)
			
		case protocol.SEND_CLIENT_ID_CMD:
			err = pp.procClientID(&cmd, session)
			if err != nil {
				log.Error("error:", err)
				return err
			}
			
		case protocol.SEND_MESSAGE_P2P_CMD:
			err = pp.procSendMessageP2P(&cmd, session)
			if err != nil {
				log.Error("error:", err)
				return err
			}
			
		case protocol.ROUTE_MESSAGE_P2P_CMD:
			err = pp.procRouteMessageP2P(&cmd, session)
			if err != nil {
				log.Error("error:", err)
				return err
			}
			
//		case protocol.CREATE_TOPIC_CMD:
//			err = pp.procCreateTopic(&cmd, session)
//			if err != nil {
//				log.Error("error:", err)
//				return err
//			}
//		case protocol.JOIN_TOPIC_CMD:
//			err = pp.procJoinTopic(&cmd, session)
//			if err != nil {
//				log.Error("error:", err)
//				return err
//			}
//		case protocol.SEND_MESSAGE_TOPIC_CMD:
//			err = pp.procSendMessageTopic(&cmd, session)
//			if err != nil {
//				log.Error("error:", err)
//				return err
//			}

		// p2p ack
		case protocol.P2P_ACK_CMD:
			err = pp.procP2pAck(&cmd, session)
			if err != nil {
				log.Error("error:", err)
				return err
			}
		}

	return err
}
