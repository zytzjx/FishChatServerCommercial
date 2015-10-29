package main

import (
	"flag"
	"fmt"
	// "goProject/libnet"
	"goProject/log"
	// "goProject/protocol"
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
	fmt.Printf("monitor version %s  \n", VERSION)
}

var InputConfFile = flag.String("conf_file", "monitor.json", "input conf file name")

func main() {

	c := make(chan int)

	version()
	fmt.Printf("built on %s\n", BuildTime())
	flag.Parse()
	cfg := NewMonitorConfig(*InputConfFile)
	err := cfg.LoadConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}

	rs := NewMonitor(cfg)

	InitContainer()

	//TODO not use go
	go rs.subscribeChannels()

	go rs.startWebServer(cfg.Listen)

	for {
		select {
		case <-c:
			break
		}
	}
}
