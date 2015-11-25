package main

import (
	"encoding/json"
	"goProject/log"
	"os"
	"time"
)

type MonitorConfig struct {
	configfile            string
	TransportProtocols    string
	LocalIP               string
	Listen                string
	LogFile               string
	MsgServerList         []string
	ScanDeadServerTimeout time.Duration
	HeartBeatTime         time.Duration
	Expire                time.Duration
	RefreshServerListTime time.Duration
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

func NewMonitorConfig(configfile string) *MonitorConfig {
	return &MonitorConfig{
		configfile: configfile,
	}
}

func (self *MonitorConfig) LoadConfig() error {
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
