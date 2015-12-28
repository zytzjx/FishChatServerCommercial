package main

import (
	"goProject/log"
	"goProject/storage/mongo_store"
	"net/http"
)

type Server struct {
	Host string
	Port string
	Db   *mongo_store.MongoStore
}

func NewServer(c *Config) *Server {
	return &Server{
		Host: c.Host,
		Port: c.Port,
		Db:   mongo_store.NewMongoStore(c.Mongo.Addr, c.Mongo.Port, c.Mongo.User, c.Mongo.Password),
	}
}

func (self *Server) Init() {
	var (
		h *handle
	)

	h = NewHandle(self.Db)

	log.Infof("server start: %s: %s", self.Host, self.Port)
	http.HandleFunc("/", h.Route)
	http.ListenAndServe(":"+self.Port, nil)
}
