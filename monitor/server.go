package main

import (
	// "goProject/base"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/storage/mongo_store"
	"sync"
	// "time"
	"net/http"
)

type Monitor struct {
	cfg                       *MonitorConfig
	connectedMsgServerList    []string
	disConnectedMsgServerList []string
	msgServerClientMap        map[string]*libnet.Session
	mongoStore                *mongo_store.MongoStore
	msgServerMutex            sync.Mutex
}

func NewMonitor(cfg *MonitorConfig) *Monitor {
	return &Monitor{
		cfg:                cfg,
		msgServerClientMap: make(map[string]*libnet.Session),
		mongoStore:         mongo_store.NewMongoStore(cfg.Mongo.Addr, cfg.Mongo.Port, cfg.Mongo.User, cfg.Mongo.Password),
	}
}

//连接msgServer
func (self *Monitor) connectServer(ms string) (*libnet.Session, error) {
	client, err := libnet.Connect("tcp", ms, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, err
}

//连接未连接的MsgServerList
func (self *Monitor) connectDisconnectedMsgServerList() {
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
			msgServerClient.State = ms

			self.msgServerMutex.Unlock()

			// go self.heartBeatWithMsgServer(msgServerClient, ms)
			go func() {
				for {
					var msg protocol.CmdMonitor
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

// //扫描没有连接的Server
// func (self *Monitor) scanDisconnectServer() {
// 	log.Info("scanDisconnectServer")
// 	timer := time.NewTicker(self.cfg.ScanDeadServerTimeout * time.Second)
// 	ttl := time.After(10 * time.Second)
// 	for {
// 		select {
// 		case <-timer.C:
// 			go self.connectDisconnectedMsgServerList()
// 		case <-ttl:
// 			break
// 		}
// 	}
// }

//解析Server过来的命令
func (self *Monitor) parseProtocol(msg protocol.CmdMonitor, sc *libnet.Session) error {
	var err error

	pp := NewProtoProc(self)
	err = pp.procMonitorMsg(msg, sc)
	// log.Info(msg)

	return err
}

// //保持到MsgServer的心跳
// func (self *Monitor) heartBeatWithMsgServer(msgServerClient *libnet.Session, ms string) {
// 	log.Info("heartBeat with ", ms)

// 	timer := time.NewTicker(self.cfg.HeartBeatTime * time.Second)
// 	ttl := time.After(self.cfg.Expire * time.Second)
// xf:
// 	for {
// 		select {
// 		case <-timer.C:
// 			cmd := protocol.NewCmdSimple(protocol.SEND_PING_CMD)
// 			err := msgServerClient.Send(cmd)
// 			if err != nil {
// 				log.Info(ms, " : disconnect")

// 				self.msgServerMutex.Lock()

// 				self.connectedMsgServerList = common.DeleteChild(self.connectedMsgServerList, ms)
// 				self.disConnectedMsgServerList = append(self.disConnectedMsgServerList, ms)
// 				delete(self.msgServerClientMap, ms)

// 				self.msgServerMutex.Unlock()

// 				break xf
// 			}
// 		case <-ttl:
// 			break
// 		}
// 	}
// }

//开始订阅Channels
func (self *Monitor) subscribeChannels() error {
	log.Info("monitor start to subscribeChannels")
	var err error

	self.disConnectedMsgServerList = self.cfg.MsgServerList

	go self.connectDisconnectedMsgServerList()

	// go self.scanDisconnectServer()

	return err
}

//接通MsgServer后执行的命令
func (self *Monitor) connectMsgServerCommands(msgServerClient *libnet.Session) error {
	var err error
	cmd := protocol.NewCmdSimple(protocol.SUBSCRIBE_CHANNEL_CMD)
	cmd.AddArg(protocol.SYSCTRL_MONITOR)
	cmd.AddArg(self.cfg.Listen)

	err = msgServerClient.Send(cmd)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

//webServer
func (self *Monitor) startWebServer(port string) {
	http.HandleFunc("/", PageIndex)
	http.HandleFunc("/api/msg_server", ApiMsgServer)
	http.ListenAndServe(":"+port, nil)
}
