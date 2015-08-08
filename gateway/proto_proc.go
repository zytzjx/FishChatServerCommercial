package main

import (
	"flag"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type ProtoProc struct {
	gateway *Gateway
}

func NewProtoProc(gateway *Gateway) *ProtoProc {
	return &ProtoProc{
		gateway: gateway,
	}
}

func (self *ProtoProc) procReqMsgServer(cmd *protocol.CmdSimple, session *libnet.Session) error {
	log.Info("procReqMsgServer")
	var err error
	msgServer := common.SelectServer(self.gateway.cfg.MsgServerList, self.gateway.cfg.MsgServerNum)

	resp := protocol.NewCmdSimple(protocol.SELECT_MSG_SERVER_FOR_CLIENT_CMD)
	resp.AddArg(msgServer)

	log.Info("Resp | ", resp)

	if session != nil {
		if err = session.Send(resp); err != nil {
			log.Error(err.Error())
		}
		session.Close()

		log.Info("client ", session.Conn().RemoteAddr().String(), " | close")
	}
	return err
}
