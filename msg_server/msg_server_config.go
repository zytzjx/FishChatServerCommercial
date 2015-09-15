package main

import (
	"encoding/json"
	"goProject/log"
	"os"
	"time"
)

type MsgServerConfig struct {
	configfile               string
	LocalIP                  string
	TransportProtocols       string
	Listen                   string
	LogFile                  string
	ScanDeadSessionTimeout   time.Duration
	ScanTimeoutAck           time.Duration
	Expire                   time.Duration
	MonitorBeatTime          time.Duration
	SessionManagerServerList []string
	Redis                    struct {
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

func NewMsgServerConfig(configfile string) *MsgServerConfig {
	return &MsgServerConfig{
		configfile: configfile,
	}
}

func (self *MsgServerConfig) LoadConfig() error {
	file, err := os.Open(self.configfile)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	err = dec.Decode(&self)
	if err != nil {
		return err
	}
	return nil
}

func (self *MsgServerConfig) DumpConfig() {
	//fmt.Printf("Mode: %s\nListen: %s\nServer: %s\nLogfile: %s\n",
	//cfg.Mode, cfg.Listen, cfg.Server, cfg.Logfile)
}
