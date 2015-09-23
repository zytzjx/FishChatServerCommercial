
package main

import (
	"flag"
	"goProject/log"
	"goProject/libnet"
	"goProject/protocol"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type Gateway struct {
	cfg     *GatewayConfig
	server  *libnet.Server
}

func NewGateway(cfg *GatewayConfig) *Gateway {
	return &Gateway {
		cfg    : cfg,
		server : new(libnet.Server),
	}
}

func (self *Gateway)parseProtocol(cmd protocol.CmdSimple, session *libnet.Session) error {	
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
