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

const VERSION string = "0.0.1"

func version() {
	fmt.Printf("gateway version %s   \n", VERSION)
}

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

var InputConfFile = flag.String("conf_file", "gateway.json", "input conf file name")

func handleSession(gw *Gateway, session *libnet.Session) {
	for {
		var msg protocol.CmdSimple
		if err := session.Receive(&msg); err != nil {
			break
		}

		err := gw.parseProtocol(msg, session)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func main() {
	version()
	fmt.Printf("built on %s\n", BuildTime())
	flag.Parse()
	cfg := NewGatewayConfig(*InputConfFile)
	err := cfg.LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}

	gw := NewGateway(cfg)

	//gw.server, err = libnet.Serve(cfg.TransportProtocols, cfg.Listen, libnet.Json())
	gw.server, err = libnet.Serve(cfg.TransportProtocols, cfg.Listen, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		panic(err)
	}

	go gw.serviceDiscovery()

	log.Info("gateway server start:", gw.server.Listener().Addr().String())
	for {
		session, err := gw.server.Accept()
		if err != nil {
			break
		}

		go handleSession(gw, session)
	}
}
