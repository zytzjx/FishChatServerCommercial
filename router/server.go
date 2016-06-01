package main

import (
	// "goProject/base"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	"sync"
	"time"
)

type Router struct {
	server                        *libnet.Server
	cfg                           *RouterConfig
	connectedMsgServerList        []string
	disConnectedMsgServerList     []string
	msgServerClientMap            map[string]*libnet.Session
	connectedBrotherServerList    []string
	disConnectedBrotherServerList []string
	brotherServerClientMap        map[string]*libnet.Session //收消息的连接集合
	brotherServerMap              map[string]*libnet.Session //发消息的连接集合
	otherMsgServerMap             map[string][]string
	mongoStore                    *mongo_store.MongoStore
	msgServerMutex                sync.Mutex
	brotherServerMutex            sync.Mutex
	otherMsgServerMutex           sync.Mutex
}

func NewRouter(cfg *RouterConfig) *Router {
	return &Router{
		cfg:                    cfg,
		msgServerClientMap:     make(map[string]*libnet.Session),
		brotherServerClientMap: make(map[string]*libnet.Session),
		brotherServerMap:       make(map[string]*libnet.Session),
		otherMsgServerMap:      make(map[string][]string),
		mongoStore:             mongo_store.NewMongoStore(cfg.Mongo.Addr, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password),
	}
}

//连接msgServer
func (self *Router) connectServer(ms string) (*libnet.Session, error) {
	client, err := libnet.Connect("tcp", ms, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		// log.Error(err.Error())
		return nil, err
	}

	return client, err
}

//连接未连接的MsgServerList
func (self *Router) connectDisconnectedMsgServerList() {
	for _, ms := range self.disConnectedMsgServerList {
		msgServerClient, err := self.connectServer(ms)
		if err == nil {
			log.Info(ms, " : connected")

			err = self.connectMsgServerCommands(msgServerClient)
			if err != nil {
				log.Error(err.Error())
			}

			self.msgServerMutex.Lock()

			self.disConnectedMsgServerList = common.DeleteChild(self.disConnectedMsgServerList, ms)
			self.connectedMsgServerList = append(self.connectedMsgServerList, ms)
			self.msgServerClientMap[ms] = msgServerClient

			self.msgServerMutex.Unlock()

			go self.heartBeatWithMsgServer(msgServerClient, ms)
			go func() {
				for {
					var msg protocol.CmdSimple
					if err := msgServerClient.Receive(&msg); err != nil {
						break
					}

					err := self.parseProtocol(msg, msgServerClient)
					if err != nil {
						log.Error(err.Error())
					}
				}
			}()
		}
	}
}

//连接未连接的BrotherList
func (self *Router) connectDisconnectedBrotherServerList() {
	for _, rs := range self.disConnectedBrotherServerList {
		brotherServerClient, err := self.connectServer(rs)
		if err == nil {
			log.Info(rs, " : connected")

			err = self.connectBrotherServerCommands(brotherServerClient)
			if err != nil {
				log.Error(err.Error())
			}

			self.brotherServerMutex.Lock()

			self.disConnectedBrotherServerList = common.DeleteChild(self.disConnectedBrotherServerList, rs)
			self.connectedBrotherServerList = append(self.connectedBrotherServerList, rs)
			self.brotherServerClientMap[rs] = brotherServerClient

			self.brotherServerMutex.Unlock()

			go self.heartBeatWithBrotherServer(brotherServerClient, rs)
			go func() {
				for {
					var msg protocol.CmdSimple
					if err := brotherServerClient.Receive(&msg); err != nil {
						break
					}

					err := self.parseProtocol(msg, brotherServerClient)
					if err != nil {
						log.Error(err.Error())
					}
				}
			}()
		}
	}
}

//获取Brother管理的MsgServer列表
func (self *Router) getBrotherManageServers() {
	// log.Info("getBrotherManageServers")
	result := self.mongoStore.ReadRouterManagerMsgServers(mongo_store.DATA_BASE_NAME, mongo_store.KV_COLLECTION, self.cfg.BrotherServerList)
	if result != nil {

		self.otherMsgServerMutex.Lock()

		for _, v := range result {
			self.otherMsgServerMap[v.Key] = v.Value
		}

		self.otherMsgServerMutex.Unlock()

	}
}

//扫描没有连接的Server
func (self *Router) scanDisconnectServer() {
	log.Info("scanDisconnectServer")
	timer := time.NewTicker(self.cfg.ScanDeadServerTimeout * time.Second)
	ttl := time.After(10 * time.Second)
	for {
		select {
		case <-timer.C:
			go self.connectDisconnectedMsgServerList()
			go self.connectDisconnectedBrotherServerList()
		case <-ttl:
			break
		}
	}
}

//刷新获取连接
func (self *Router) refreshManageServers() {
	log.Info("refreshManageServers")
	timer := time.NewTicker(self.cfg.RefreshServerListTime * time.Second)
	ttl := time.After(10 * time.Second)
	for {
		select {
		case <-timer.C:
			go self.getBrotherManageServers()
		case <-ttl:
			break
		}
	}
}

//解析Server过来的命令
func (self *Router) parseProtocol(msg protocol.CmdSimple, sc *libnet.Session) error {
	var err error

	pp := NewProtoProc(self)

	switch msg.GetCmdName() {
	case protocol.SUBSCRIBE_CHANNEL_CMD:
		err = pp.procSubscribeChannel(&msg, sc)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	case protocol.ROUTE_MSG_CMD:
		err = pp.procRouteMsg(&msg, sc)
		if err != nil {
			log.Warning(err.Error())
		}
	case protocol.SEND_PING_CMD:
	case protocol.RESP_PONG_CMD:
	default:
		log.Info(msg)
	}

	return err
}

//保持到MsgServer的心跳
func (self *Router) heartBeatWithMsgServer(msgServerClient *libnet.Session, ms string) {
	log.Info("heartBeat with ", ms)

	timer := time.NewTicker(self.cfg.HeartBeatTime * time.Second)
	ttl := time.After(self.cfg.Expire * time.Second)
xf:
	for {
		select {
		case <-timer.C:
			cmd := protocol.NewCmdSimple(protocol.SEND_PING_CMD)
			err := msgServerClient.Send(cmd)
			if err != nil {
				log.Info(ms, " : disconnect")

				self.msgServerMutex.Lock()

				self.connectedMsgServerList = common.DeleteChild(self.connectedMsgServerList, ms)
				self.disConnectedMsgServerList = append(self.disConnectedMsgServerList, ms)
				delete(self.msgServerClientMap, ms)

				self.msgServerMutex.Unlock()

				break xf
			}
		case <-ttl:
			break
		}
	}
}

//保持到BrotherServer的心跳
func (self *Router) heartBeatWithBrotherServer(brotherServerClient *libnet.Session, rs string) {
	log.Info("heartBeat with ", rs)

	timer := time.NewTicker(self.cfg.HeartBeatTime * time.Second)
	ttl := time.After(self.cfg.Expire * time.Second)
xe:
	for {
		select {
		case <-timer.C:
			cmd := protocol.NewCmdSimple(protocol.SEND_PING_CMD)
			err := brotherServerClient.Send(cmd)
			if err != nil {
				log.Info(rs, " : disconnect")

				self.brotherServerMutex.Lock()

				self.connectedBrotherServerList = common.DeleteChild(self.connectedBrotherServerList, rs)
				self.disConnectedBrotherServerList = append(self.disConnectedBrotherServerList, rs)
				delete(self.brotherServerClientMap, rs)

				self.brotherServerMutex.Unlock()

				break xe
			}
		case <-ttl:
			break
		}
	}
}

//开始订阅Channels
func (self *Router) subscribeChannels() error {
	log.Info("route start to subscribeChannels")
	var err error

	kvStoreData := mongo_store.KVData{"routerManageMsgServer", self.cfg.LocalIP + ":" + self.cfg.Listen, self.cfg.MsgServerList}

	//更新小弟列表到数据库
	err = self.mongoStore.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.KV_COLLECTION, &kvStoreData)
	if err != nil {
		log.Error(err.Error())
	}

	self.disConnectedMsgServerList = self.cfg.MsgServerList
	self.disConnectedBrotherServerList = self.cfg.BrotherServerList

	go self.connectDisconnectedMsgServerList()
	go self.connectDisconnectedBrotherServerList()
	self.getBrotherManageServers()

	go self.scanDisconnectServer()
	go self.refreshManageServers()

	return err
}

//接通MsgServer后执行的命令
func (self *Router) connectMsgServerCommands(msgServerClient *libnet.Session) error {
	var err error
	cmd := protocol.NewCmdSimple(protocol.SUBSCRIBE_CHANNEL_CMD)
	cmd.AddArg(protocol.SYSCTRL_SEND)
	cmd.AddArg(self.cfg.Listen)

	err = msgServerClient.Send(cmd)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	cmd = protocol.NewCmdSimple(protocol.SUBSCRIBE_CHANNEL_CMD)
	cmd.AddArg(protocol.SYSCTRL_TOPIC_SYNC)
	cmd.AddArg(self.cfg.Listen)

	err = msgServerClient.Send(cmd)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//接通BrotherServer执行命令
func (self *Router) connectBrotherServerCommands(brotherServerClient *libnet.Session) error {
	var err error
	cmd := protocol.NewCmdSimple(protocol.SUBSCRIBE_CHANNEL_CMD)
	// cmd.AddArg(protocol.SYSCTRL_SEND)
	cmd.AddArg(self.cfg.LocalIP + ":" + self.cfg.Listen)

	err = brotherServerClient.Send(cmd)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}
