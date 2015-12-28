package main

import (
	"goProject/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	Host  string `yaml: "host"`
	Port  string `yaml: "port"`
	Mongo struct {
		Addr     string `yaml: "addr"`
		Port     string `yaml: "port"`
		User     string `yaml: "user"`
		Password string `yaml: "password"`
	}
	file string
	f    *os.File
}

func NewConfig(file string) (c *Config, err error) {
	var data []byte
	c = &Config{}
	c.file = file
	if c.f, err = os.OpenFile(file, os.O_RDONLY, 0664); err != nil {
		log.Errorf("os.OpenFile(\"%s\") error(%v)", file, err)
		return
	}
	if data, err = ioutil.ReadAll(c.f); err != nil {
		log.Errorf("ioutil.ReadAll(\"%s\") error(%v)", file, err)
		goto failed
	}
	err = yaml.Unmarshal(data, c)
failed:
	c.f.Close()
	return
}
