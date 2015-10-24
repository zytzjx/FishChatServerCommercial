package main

import (
	"fmt"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"net"
	"sync"
	"time"
)

type MonitorBeat struct {
	name      string
	session   *libnet.Session
	mu        sync.Mutex
	timeout   time.Duration
	expire    time.Duration
	fails     uint64
	threshold uint64
}

func NewMonitorBeat(name string, timeout time.Duration, expire time.Duration, limit uint64) *MonitorBeat {
	return &MonitorBeat{
		name:      name,
		timeout:   timeout,
		expire:    expire,
		threshold: limit,
	}
}

func (self *MonitorBeat) ResetFailures() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.fails = 0
}

func (self *MonitorBeat) ChangeThreshold(thres uint64) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.threshold = thres
}

func (self *MonitorBeat) Beat(c *libnet.Channel, data *protocol.CmdMonitor) {
	timer := time.NewTicker(self.timeout * time.Second)
	//ttl := time.After(self.expire * time.Second)
	//for {
	select {
	case <-timer.C:
		go func() {
			err := c.Broadcast(data)
			if err != nil {
				log.Error(err.Error())
				//return err
			}
		}()
		//case <-ttl:
		//break
	}
	//}

	//return nil
}

func (self *MonitorBeat) Receive() {
	timeout := time.After(self.timeout)
	for {
		select {
		case <-timeout:
			self.fails = self.fails + 1
			if self.fails > self.threshold {
				return
			}
		}
	}
}

// TODO : no use
func getHostIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		fmt.Println(addr.String())
	}
}

//---------------------------------------------------------------------------
// 命令映射表,用于执行send_message_p2p,send_notify_p2p等相应的命令
//---------------------------------------------------------------------------
type CommendMapped struct {
	RespCmd    string
	ReceiveCmd string
	RouterCmd  string
}

type CommendMappedMap map[string]CommendMapped

var NCommendMappedMap CommendMappedMap

func InitCommendMapped() {
	NCommendMappedMap = make(CommendMappedMap)

	NCommendMappedMap[protocol.SEND_MESSAGE_P2P_CMD] = CommendMapped{
		protocol.RESP_MESSAGE_P2P_CMD,
		protocol.RECEIVE_MESSAGE_P2P_CMD,
		protocol.ROUTE_MESSAGE_P2P_CMD,
	}
	NCommendMappedMap[protocol.SEND_NOTIFY_P2P_CMD] = CommendMapped{
		protocol.RESP_NOTIFY_P2P_CMD,
		protocol.RECEIVE_NOTIFY_P2P_CMD,
		protocol.ROUTE_NOTIFY_P2P_CMD,
	}
	NCommendMappedMap[protocol.SEND_MESSAGE_TOPIC_CMD] = CommendMapped{
		protocol.RESP_MESSAGE_TOPIC_CMD,
		protocol.RECEIVE_MESSAGE_TOPIC_CMD,
		protocol.ROUTE_MESSAGE_TOPIC_CMD,
	}
	NCommendMappedMap[protocol.SEND_NOTIFY_TOPIC_CMD] = CommendMapped{
		protocol.RESP_NOTIFY_TOPIC_CMD,
		protocol.RECEIVE_NOTIFY_TOPIC_CMD,
		protocol.ROUTE_NOTIFY_TOPIC_CMD,
	}
}
