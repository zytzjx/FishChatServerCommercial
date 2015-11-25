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
	// smsg.Repo = []string{"hello", "aabbcc", "test"}

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

	fmt.Print("Input my id :")
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

		//获取操作
		fmt.Print("Command (create,join,invite,leave,list,member,send,logout) : ")
		if _, err = fmt.Scanf("%s\n", &input); err != nil {
			log.Error(err.Error())
		}

		switch input {
		case "create":
			smsg = protocol.NewCmdSimple(protocol.SEND_CREATE_TOPIC_CMD)
			fmt.Print("Please input topic name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)
			// smsg.AddArg(myID)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "join":
			smsg = protocol.NewCmdSimple(protocol.SEND_JOIN_TOPIC_CMD)
			fmt.Print("Please input topic name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)
			// smsg.AddArg(myID)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "invite":
			smsg = protocol.NewCmdSimple(protocol.SEND_INVITE_TOPIC_CMD)
			fmt.Print("Please input topic name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)

			fmt.Print("Please input invite friends name.")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}
			smsg.AddArg(input)
			fmt.Print("Please input invite friends name.")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}
			smsg.AddArg(input)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}

		case "send":
			smsg = protocol.NewCmdSimple(protocol.SEND_MESSAGE_TOPIC_CMD)
			fmt.Print("Send message to topic : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			topicID := input

			fmt.Print("input msg :")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}
			msg := input

			smsg.AddArg(msg)
			smsg.AddArg(topicID)
			// smsg.AddArg(myID)
			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "leave":
			smsg = protocol.NewCmdSimple(protocol.SEND_LEAVE_TOPIC_CMD)
			fmt.Print("Please input topic name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)
			// smsg.AddArg(myID)

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "list":
			smsg = protocol.NewCmdSimple(protocol.SEND_LIST_TOPIC_CMD)
			smsg.Repo = "hello world"

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
			}
		case "member":
			smsg = protocol.NewCmdSimple(protocol.SEND_TOPIC_MEMBERS_LIST_CMD)
			fmt.Print("Please input topic name : ")
			if _, err = fmt.Scanf("%s\n", &input); err != nil {
				log.Error(err.Error())
			}

			smsg.AddArg(input)
			smsg.Repo = "hello world"

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

	fmt.Println(cmd)
	switch cmd.GetCmdName() {
	case protocol.RECEIVE_MESSAGE_TOPIC_CMD:
		newCmd := protocol.NewCmdSimple(protocol.TOPIC_ACK_CMD)
		// newCmd.AddArg(myID)
		newCmd.AddArg(cmd.GetArgs()[4])
		if session != nil {
			err = session.Send(newCmd)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		}
	case protocol.RESP_PONG_CMD:
	default:

	}

	return err
}
