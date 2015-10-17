

package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"goProject/log"
	"goProject/libnet"
	"goProject/protocol"
)

type MonitorBeat struct {
	name       string
	session    *libnet.Session
	mu         sync.Mutex
	timeout    time.Duration
	expire     time.Duration
	fails      uint64
	threshold  uint64
}

func NewMonitorBeat(name string, timeout time.Duration, expire time.Duration, limit uint64) *MonitorBeat {
	return &MonitorBeat {
		name      : name,
		timeout   : timeout,
		expire    : expire,
		threshold : limit,
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


