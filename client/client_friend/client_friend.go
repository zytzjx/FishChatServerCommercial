package main

import (
	"flag"
	"fmt"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
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

	fmt.Print("input my id :")
	var myID string
	if _, err := fmt.Scanf("%s\n", &myID); err != nil {
		log.Error(err.Error())
	}

	smsg.AddArg(myID)

	//告诉服务器我的ID
	err = msgServerClient.Send(smsg)
	if err != nil {
		log.Error(err.Error())
	}

	go heartBeat(cfg, msgServerClient)

	var input string

	go func() {
		for {
			if err := msgServerClient.Receive(&rmsg); err != nil {
				break
			}
			parseProtocol(rmsg, msgServerClient, myID)
		}
	}()

	for {
		//获取操作
		fmt.Print("Command (add,newadd,del,list,logout) : ")
		if _, err = fmt.Scanf("%s\n", &input); err != nil {
			log.Error(err.Error())
		}

		switch input {
		case "add":
			smsg = protocol.NewCmdSimple(protocol.SEND_ADD_FRIEND_CMD)
			fmt.Print("Please input friend name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "newadd":
			smsg = protocol.NewCmdSimple(protocol.SEND_ASK_CMD)
			fmt.Print("Please input friend name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(protocol.SEND_ASK_CMD_TYPE_ADD_FRIEND)
			smsg.AddArg(input)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "del":
			smsg = protocol.NewCmdSimple(protocol.SEND_DEL_FRIEND_CMD)
			fmt.Print("Please input friend name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "list":
			smsg = protocol.NewCmdSimple(protocol.SEND_VIEW_FRIENDS_CMD)
			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "logout":
			smsg = protocol.NewCmdSimple(protocol.SEND_LOGOUT_CMD)
			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		}

	}

}

func parseProtocol(cmd protocol.CmdResponse, session *libnet.Session, myID string) error {
	var err error

	switch cmd.GetCmdName() {
	case protocol.RESP_PONG_CMD:
	case protocol.RECEIVE_ASK_CMD:
		log.Info("agree friend ask.")
		msg := protocol.NewCmdSimple(protocol.SEND_REACT_CMD)

		msg.AddArg(protocol.SEND_REACT_CMD_AGREE)
		msg.AddArg(cmd.GetArgs()[3])
		err = session.Send(msg)
		if err != nil {
			log.Error(err.Error())
		}
	default:
		log.Info(cmd)
	}

	return err
}
