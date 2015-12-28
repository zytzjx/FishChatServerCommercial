package main

import (
	"flag"
	"fmt"
	"goProject/libnet"
	"goProject/log"
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

const VERSION string = "0.1.6"

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

func version() {
	fmt.Printf("router version %s  \n", VERSION)
}

var InputConfFile = flag.String("conf_file", "router.json", "input conf file name")

func handleSession(rs *Router, session *libnet.Session) {
	// log.Info("a new client ", session.Conn().RemoteAddr().String(), " | come in")
	for {
		var msg protocol.CmdSimple
		if err := session.Receive(&msg); err != nil {
			break
		}

		err := rs.parseProtocol(msg, session)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func main() {
	version()
	fmt.Printf("built on %s\n", BuildTime())
	flag.Parse()
	cfg := NewRouterConfig(*InputConfFile)
	err := cfg.LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}

	rs := NewRouter(cfg)

	rs.server, err = libnet.Serve(cfg.TransportProtocols, ":"+cfg.Listen, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("router server start: ", rs.server.Listener().Addr().String())

	//TODO not use go
	go rs.subscribeChannels()

	for {
		session, err := rs.server.Accept()
		if err != nil {
			log.Error(err.Error())
			break
		}

		handleSession(rs, session)
	}

}
