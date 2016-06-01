package main

import (
	"flag"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/service_discovery"
	"sync"
	"time"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type Gateway struct {
	cfg                *GatewayConfig
	server             *libnet.Server
	master             *service_discovery.Master
	msgServerList      []MsgServerInfo
	msgServerListMutex sync.Mutex
}

type MsgServerInfo struct {
	Ip         string
	CPU        int
	SessionNum uint64
}

func NewGateway(cfg *GatewayConfig) *Gateway {
	return &Gateway{
		cfg:    cfg,
		server: new(libnet.Server),
		master: service_discovery.NewMaster([]string{cfg.EtcdServer}),
	}
}

func (self *Gateway) serviceDiscovery() {
	log.Info("serviceDiscovery")
	timer := time.NewTicker(self.cfg.ServiceDiscoveryTimeout * time.Second)
	ttl := time.After(10 * time.Second)
	for {
		select {
		case <-timer.C:
			self.msgServerListMutex.Lock()
			self.msgServerList = []MsgServerInfo{}
			for _, v := range self.master.Members {
				if v.InGroup == true {
					self.msgServerList = append(self.msgServerList, MsgServerInfo{v.IP, v.CPU, v.SessionNum})
				}
			}
			self.msgServerListMutex.Unlock()
		case <-ttl:
			break
		}
	}
}

func (self *Gateway) parseProtocol(cmd protocol.CmdSimple, session *libnet.Session) error {
	var err error
	pp := NewProtoProc(self)

	switch cmd.GetCmdName() {
	case protocol.REQ_MSG_SERVER_CMD:
		err = pp.procReqMsgServer(&cmd, session)
		if err != nil {
			log.Error("error:", err)
			return err
		}
	}

	return err
}
