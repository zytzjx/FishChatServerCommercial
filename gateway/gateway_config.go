

package main

import (
	"os"
	"encoding/json"
	"goProject/log"
)

type GatewayConfig struct {
	configFile         string
	TransportProtocols string
	Listen             string
	LogFile            string
	MsgServerList      []string
	MsgServerNum       int
}

func NewGatewayConfig(configFile string) *GatewayConfig {
	return &GatewayConfig {
		configFile : configFile,
	}
}

func (self *GatewayConfig)LoadConfig() error {
	file, err := os.Open(self.configFile)
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

func (self *GatewayConfig)DumpConfig() {
	//fmt.Printf("Mode: %s\nListen: %s\nServer: %s\nLogfile: %s\n", 
	//cfg.Mode, cfg.Listen, cfg.Server, cfg.Logfile)
}
