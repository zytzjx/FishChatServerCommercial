package main

import (
	"sync"
	//"time"
	"goProject/log"
	//"goProject/base"
	"goProject/libnet"
	"goProject/protocol"
	"goProject/storage/mongo_store"
)

type Router struct {
	cfg                *RouterConfig
	msgServerClientMap map[string]*libnet.Session
	topicServerMap     map[string]string
	readMutex          sync.Mutex

	mongoStore *mongo_store.MongoStore
}

func NewRouter(cfg *RouterConfig) *Router {
	return &Router{
		cfg:                cfg,
		msgServerClientMap: make(map[string]*libnet.Session),
		topicServerMap:     make(map[string]string),

		mongoStore: mongo_store.NewMongoStore(cfg.Mongo.Addr, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password),
	}
}

func (self *Router) connectMsgServer(ms string) (*libnet.Session, error) {
	client, err := libnet.Connect("tcp", ms, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}

	return client, err
}

func (self *Router) handleMsgServerClient(msc *libnet.Session) {
	pp := NewProtoProc(self)

	for {
		var msg protocol.CmdInternal
		if err := msc.Receive(&msg); err != nil {
			break
		}
		log.Info("msg_server", msc.Conn().RemoteAddr().String(), " say: ", msg)

		switch msg.GetCmdName() {
		case protocol.SEND_MESSAGE_P2P_CMD:
			err := pp.procSendMsgP2P(&msg, msc)
			if err != nil {
				log.Warning(err.Error())
			}
		case protocol.CREATE_TOPIC_CMD:
			err := pp.procCreateTopic(&msg, msc)
			if err != nil {
				log.Warning(err.Error())
			}
		case protocol.JOIN_TOPIC_CMD:
			err := pp.procJoinTopic(&msg, msc)
			if err != nil {
				log.Warning(err.Error())
			}
		case protocol.SEND_MESSAGE_TOPIC_CMD:
			err := pp.procSendMsgTopic(&msg, msc)
			if err != nil {
				log.Warning(err.Error())
			}

		}
	}

}

func (self *Router) subscribeChannels() error {
	log.Info("route start to subscribeChannels")
	for _, ms := range self.cfg.MsgServerList {
		msgServerClient, err := self.connectMsgServer(ms)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		cmd := protocol.NewCmdSimple(protocol.SUBSCRIBE_CHANNEL_CMD)
		cmd.AddArg(protocol.SYSCTRL_SEND)
		cmd.AddArg(self.cfg.UUID)

		err = msgServerClient.Send(cmd)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		cmd = protocol.NewCmdSimple(protocol.SUBSCRIBE_CHANNEL_CMD)
		cmd.AddArg(protocol.SYSCTRL_TOPIC_SYNC)
		cmd.AddArg(self.cfg.UUID)

		err = msgServerClient.Send(cmd)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		self.msgServerClientMap[ms] = msgServerClient
	}

	for _, msc := range self.msgServerClientMap {
		go self.handleMsgServerClient(msc)
	}
	return nil
}
