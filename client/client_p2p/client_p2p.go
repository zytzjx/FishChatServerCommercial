package main

import (
	"fmt"
	"flag"
	"goProject/log"
	"goProject/libnet"
	"goProject/common"
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
	
	var rmsg protocol.CmdSimple
	
	if err := gatewayClient.Receive(&rmsg); err != nil {
		log.Error(err.Error())
	}
	log.Info(rmsg)
	
	msgServerClient, err := libnet.Connect("tcp", rmsg.GetArgs()[0], libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		panic(err)
	}
	
	smsg = protocol.NewCmdSimple(protocol.SEND_CLIENT_ID_CMD)
	
	fmt.Println("input my id :")
	var myID string
	if _, err := fmt.Scanf("%s\n", &myID); err != nil {
		log.Error(err.Error())
	}
	
	smsg.AddArg(myID)
	
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
			fmt.Printf("%s\n", rmsg)
		}
	}()
	
	for {
		smsg = protocol.NewCmdSimple(protocol.SEND_MESSAGE_P2P_CMD)
		
		fmt.Println("send the id you want to talk :")
		if _, err = fmt.Scanf("%s\n", &input); err != nil {
			log.Error(err.Error())
		}
		
		smsg.AddArg(input)
		
		fmt.Println("input msg :")
		if _, err = fmt.Scanf("%s\n", &input); err != nil {
			log.Error(err.Error())
		}
		
		smsg.AddArg(input)
		
		smsg.AddArg(myID)
		
		err = msgServerClient.Send(smsg)
		if err != nil {
			log.Error(err.Error())
		}
	}
	
}
