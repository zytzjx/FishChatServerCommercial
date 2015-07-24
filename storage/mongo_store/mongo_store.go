

package mongo_store

import (
	"sync"
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"goProject/log"
)

type MongoStoreOptions struct {

}

type MongoStore struct {
	opts            *MongoStoreOptions
	session         *mgo.Session
	
	rwMutex         sync.Mutex
}

func NewMongoStore(ip string, port string, user string, password string) *MongoStore {
	var url string
	if user == "" && password == "" {
		url = ip + port
	} else {
		url = user + ":" + password + "@" + ip + port
	}
	
	log.Info("connect to mongo : " , url)
	maxWait := time.Duration(5 * time.Second)
	session, err := mgo.DialWithTimeout(url, maxWait)
	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		panic(err)
	}
	return &MongoStore {
		session : session,
	}
}

func (self *MongoStore)Init() {
	//self.session.DB("im").C("client_info")
	
}

func (self *MongoStore)Update(db string, c string, data interface{}) error {
	log.Info("MongoStore Update")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	
	op := self.session.DB(db).C(c)
	
	switch data.(type) {
		case *SessionStoreData:
			cid := data.(*SessionStoreData).ClientID
			log.Info("cid : " , cid)
			_, err = op.Upsert(bson.M{"ClientID": cid}, data.(*SessionStoreData))
			if err != nil {
				log.Error(err.Error())
				return err
			}
		
		case *TopicStoreData:
			topicName := data.(*TopicStoreData).TopicName
			log.Info("topicName : " , topicName)
			_, err = op.Upsert(bson.M{"TopicName": topicName}, data.(*TopicStoreData))
			if err != nil {
				log.Error(err.Error())
				return err
			}
	}
	
	return err
}

func (self *MongoStore)GetSessionFromCid(db string, c string, cid string) (*SessionStoreData, error) {
	log.Info("MongoStore GetSessionFromCid")
	log.Info(cid)
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()	
	
	op := self.session.DB(db).C(c)
	
	var result *SessionStoreData
	//var result interface{}
	
	err = op.Find(bson.M{"ClientID": cid}).One(&result)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	
	return result, nil
}


func (self *MongoStore)Close() {
	self.session.Close()
}
