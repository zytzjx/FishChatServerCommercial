

package main

import (
	"flag"
	"fmt"
	"goProject/log"
	"goProject/libnet"
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

const VERSION string = "0.10"

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

func version() {
	fmt.Printf("router version %s  \n", VERSION)
}

var InputConfFile = flag.String("conf_file", "router.json", "input conf file name")   

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
	
	server, err := libnet.Serve(cfg.TransportProtocols, cfg.Listen, libnet.Packet(libnet.Uint16BE, libnet.Json()))
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("router server start: ", server.Listener().Addr().String())
	
	r := NewRouter(cfg)
	//TODO not use go
	go r.subscribeChannels()
	
	for {
		_, err := server.Accept()
		if err != nil {
			log.Error(err.Error())
			break
		}
	}
	
}
