package main

import (
	"flag"
	"goProject/log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
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

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

func version() {
	log.Infof("msg_api version %s  \n", VERSION)
}

func startProfile(cpuprofile *string) error {
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("Ctrl+C to quit.")
		pprof.StopCPUProfile()
		os.Exit(1)
	}()

	return nil
}

func main() {
	var (
		err           error
		c             *Config
		s             *Server
		ck            chan int
		buildTime     = C.GoString(C.build_time())
		InputConfFile = flag.String("c", "msg_api.yaml", "input conf file name")
		cpuprofile    = flag.String("cpuprofile", "", "write cpu profile to file")
	)
	version()
	log.Infof("built on %s\n", buildTime)
	flag.Parse()

	if c, err = NewConfig(*InputConfFile); err != nil {
		log.Errorf("NewConfig(\"%s\") error(%v)", InputConfFile, err)
		return
	}

	if *cpuprofile != "" {
		startProfile(cpuprofile)
	}

	s = NewServer(c)
	go s.Init()

	for {
		select {
		case <-ck:
			return
		}
	}
}
