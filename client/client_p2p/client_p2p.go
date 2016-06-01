package main

import (
	"flag"
	"fmt"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	// "strconv"
)

var InputConfFile = flag.String("conf_file", "client.json", "input conf file name")

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

func heartBeat(cfg Config, msgServerClient *libnet.Session) {
	hb := common.NewHeartBeat("client", msgServerClient, cfg.HeartBeatTime, cfg.Expire, 10)
	hb.Beat()
}

func main() {
	flag.Parse()
	cfg, err := LoadConfig(*InputConfFile)
	if err != nil {
		log.Error(err.Error())
		return
	}

	gatewayClient, err := libnet.Connect("tcp", cfg.GatewayServer, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		panic(err)
	}

	//	go func() {
	//		var msg protocol.CmdSimple
	//		for {
	//			if err := session.Receive(&msg); err != nil {
	//				break
	//			}
	//			log.Info(msg)

	//		}
	//	}()

	smsg := protocol.NewCmdSimple(protocol.REQ_MSG_SERVER_CMD)

	if err = gatewayClient.Send(smsg); err != nil {
		log.Error(err.Error())
	}

	var rmsg protocol.CmdResponse

	if err := gatewayClient.Receive(&rmsg); err != nil {
		log.Error(err.Error())
	}
	log.Info(rmsg)

	msgServerClient, err := libnet.Connect("tcp", rmsg.GetArgs()[0], libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		panic(err)
	}

	smsg = protocol.NewCmdSimple(protocol.SEND_CLIENT_ID_CMD)
	// smsg = protocol.NewCmdSimple(protocol.SEND_PING_CMD)

	fmt.Print("input my id :")
	var myID string
	if _, err := fmt.Scanf("%s\n", &myID); err != nil {
		log.Error(err.Error())
	}

	// smsg = protocol.NewCmdSimple(protocol.SEND_TOKEN_CMD)
	// // smsg = protocol.NewCmdSimple(protocol.SEND_PING_CMD)

	// fmt.Print("input token :")
	// var myID string
	// if _, err := fmt.Scanf("%s\n", &myID); err != nil {
	// 	log.Error(err.Error())
	// }

	smsg.AddArg(myID)

	//告诉服务器我的ID
	err = msgServerClient.Send(smsg)
	if err != nil {
		log.Error(err.Error())
	}

	go heartBeat(cfg, msgServerClient)

	var input string

	go func() {
		//var msg string
		for {
			if err := msgServerClient.Receive(&rmsg); err != nil {
				break
			}
			// fmt.Printf("%s\n", rmsg)
			parseProtocol(rmsg, msgServerClient, myID)
		}
	}()

	for {

		fmt.Print("send the id you want to talk :")
		if _, err = fmt.Scanf("%s\n", &input); err != nil {
			log.Error(err.Error())
		}

		target := input

		fmt.Print("input msg :")
		if _, err = fmt.Scanf("%s\n", &input); err != nil {
			log.Error(err.Error())
		}
		msg := input

		// for i := 0; i < 100000; i++ {ROUTE_NOTIFY_P2P_CMD
		smsg = protocol.NewCmdSimple(protocol.SEND_MESSAGE_P2P_CMD)
		smsg.AddArg(msg)
		smsg.AddArg(target)
		log.Info(smsg)

		err = msgServerClient.Send(smsg)
		if err != nil {
			log.Error(err.Error())
		}
		// }
	}

}

func parseProtocol(cmd protocol.CmdResponse, session *libnet.Session, myID string) error {
	var err error

	switch cmd.GetCmdName() {
	case protocol.RECEIVE_MESSAGE_P2P_CMD:
		newCmd := protocol.NewCmdSimple(protocol.P2P_ACK_CMD)
		// newCmd.AddArg(myID)

		fmt.Println(cmd.GetArgs())

		newCmd.AddArg(cmd.GetArgs()[3])
		if session != nil {
			err = session.Send(newCmd)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	case protocol.RESP_LOGOUT_CMD:
		log.Info(cmd)
	case protocol.RESP_PONG_CMD:
	default:
		log.Info(cmd)
	}

	return nil
}
