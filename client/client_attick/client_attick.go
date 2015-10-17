package main

import (
	"flag"
	// "fmt"
	"goProject/common"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"math/rand"
	"strconv"
	"time"
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

func userConnnectServer(i int, answers []string, cfg Config) {

	if i%100 == 0 {
		log.Info("start session connect " + strconv.Itoa(i) + ".")
	}

	smsg := protocol.NewCmdSimple(protocol.REQ_MSG_SERVER_CMD)
	var rmsg protocol.CmdResponse
	gatewayClient, err := libnet.Connect("tcp", cfg.GatewayServer, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		log.Error(err)
		return
	}

	if err = gatewayClient.Send(smsg); err != nil {
		log.Error(err.Error())
		return
	}

	if err := gatewayClient.Receive(&rmsg); err != nil {
		log.Error(err.Error())
		return
	}

	msgServerClient, err := libnet.Connect("tcp", rmsg.GetArgs()[0], libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		log.Error(err)
		return
	}

	myId := "Attack" + strconv.Itoa(i)

	sendIdCmd := protocol.NewCmdSimple(protocol.SEND_CLIENT_ID_CMD)
	sendIdCmd.AddArg(myId)

	err = msgServerClient.Send(sendIdCmd)
	if err != nil {
		log.Error(err.Error())
		return
	}

	go heartBeat(cfg, msgServerClient)
	go func() {
		for {
			if err := msgServerClient.Receive(&rmsg); err != nil {
				break
			}
			parseProtocol(rmsg, msgServerClient)
		}
	}()
	go func() {
		for {
			smsg = protocol.NewCmdSimple(protocol.SEND_MESSAGE_P2P_CMD)
			smsg.AddArg(answers[rand.Intn(len(answers))])
			smsg.AddArg("cc")

			err = msgServerClient.Send(smsg)
			if err != nil {
				log.Error(err.Error())
				return
			}
			rt := time.Duration(rand.Int63n(60000000000))
			// log.Info(rt)
			time.Sleep(rt)
		}
	}()

}

func main() {
	flag.Parse()
	cfg, err := LoadConfig(*InputConfFile)
	if err != nil {
		log.Error(err.Error())
		return
	}

	c := make(chan int)

	answers := []string{
		"It_is_certain",
		"It_is_decidedly_so",
		"Without_a_doubt",
		"Yes_definitely",
		"You_may_rely_on_it",
		"As_I_see_it_yes",
		"Most_likely",
		"Outlook_good",
		"Yes",
		"Signs_point_to_yes",
		"Reply_hazy_try_again",
		"Ask_again_later",
		"Better_not_tell_you_now",
		"Cannot_predict_now",
		"Concentrate_and_ask_again",
		"Don't_count_on_it",
		"My_reply_is_no",
		"My_sources_say_no",
		"Outlook_not_so_good",
		"Very_doubtful",
	}

	for i := cfg.UserNumSkip * cfg.ClientNums; i < (cfg.UserNumSkip+1)*cfg.ClientNums; i++ {
		go userConnnectServer(i, answers, cfg)
		time.Sleep(time.Duration(20000000))
	}

	for {
		select {
		case <-c:
			break
		}
	}
}

func parseProtocol(cmd protocol.CmdResponse, session *libnet.Session) error {
	var err error

	switch cmd.GetCmdName() {
	case protocol.RESP_PONG_CMD:
		log.Info("PONG")
	case protocol.RESP_CLIENT_ID_CMD:
	default:
		log.Info(cmd)
	}

	return err
}
