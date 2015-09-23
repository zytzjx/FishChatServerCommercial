package main

import (
	"fmt"
	//"time"
	"flag"
	"goProject/log"
	//"goProject/base"
	"goProject/libnet"
	"goProject/protocol"
)

/*
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
const char* build_time(void) {
	static const char* psz_build_time = "["__DATE__ " " __TIME__ "]";
	return psz_build_time;
}
*/
import "C"

var (
	buildTime = C.GoString(C.build_time())
)

func BuildTime() string {
	return buildTime
}

const VERSION string = "0.0.1"

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

func version() {
	fmt.Printf("msg_server version %s  \n", VERSION)
}

var InputConfFile = flag.String("conf_file", "msg_server.json", "input conf file name")

func handleSession(ms *MsgServer, session *libnet.Session) {
	log.Info("a new client ", session.Conn().RemoteAddr().String(), " | come in")

	for {
		var msg protocol.CmdSimple
		if err := session.Receive(&msg); err != nil {
			break
		}

		err := ms.parseProtocol(msg, session)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func main() {
	version()
	fmt.Printf("built on %s\n", BuildTime())
	flag.Parse()
	cfg := NewMsgServerConfig(*InputConfFile)
	err := cfg.LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}

	ms := NewMsgServer(cfg)

	ms.server, err = libnet.Serve(cfg.TransportProtocols, cfg.Listen, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		panic(err)
	}
	log.Info("msg_server running at  ", ms.server.Listener().Addr().String())

	ms.createChannels()

	go ms.scanDeadSession()

	go ms.sendMonitorData()

	go ms.scanTimeoutAck()

	for {
		session, err := ms.server.Accept()
		if err != nil {
			break
		}

		go handleSession(ms, session)
	}
}
