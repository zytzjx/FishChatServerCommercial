package main

import (
	"encoding/json"
	"goProject/log"
	"os"
	"time"
)

type RouterConfig struct {
	configfile            string
	TransportProtocols    string
	LocalIP               string
	Listen                string
	LogFile               string
	ScanDeadServerTimeout time.Duration
	HeartBeatTime         time.Duration
	Expire                time.Duration
	RefreshServerListTime time.Duration
	MsgServerList         []string
	BrotherServerList     []string
	Redis                 struct {
		Addr           string
		Port           string
		ConnectTimeout time.Duration
		ReadTimeout    time.Duration
		WriteTimeout   time.Duration
	}
	Mongo struct {
		Addr     string
		Port     string
		User     string
		Password string
	}
}

func NewRouterConfig(configfile string) *RouterConfig {
	return &RouterConfig{
		configfile: configfile,
	}
}

func (self *RouterConfig) LoadConfig() error {
	file, err := os.Open(self.configfile)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	err = dec.Decode(&self)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (self *RouterConfig) DumpConfig() {
	//fmt.Printf("Mode: %s\nListen: %s\nServer: %s\nLogfile: %s\n",
	//cfg.Mode, cfg.Listen, cfg.Server, cfg.Logfile)
}
