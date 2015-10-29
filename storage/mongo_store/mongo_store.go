package mongo_store

import (
	"goProject/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

type MongoStoreOptions struct {
}

type MongoStore struct {
	opts    *MongoStoreOptions
	session *mgo.Session

	rwMutex sync.Mutex
}

func NewMongoStore(ip string, port string, user string, password string) *MongoStore {
	var url string
	if user == "" && password == "" {
		url = ip + port
	} else {
		url = user + ":" + password + "@" + ip + port
	}

	log.Info("connect to mongo : ", url)
	maxWait := time.Duration(5 * time.Second)
	session, err := mgo.DialWithTimeout(url, maxWait)
	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		panic(err)
	}
	return &MongoStore{
		session: session,
	}
}

func (self *MongoStore) Init() {
	//self.session.DB("im").C("client_info")
}

//新增和修改
func (self *MongoStore) Upsert(db string, c string, data interface{}) error {
	log.Info("MongoStore Upsert")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	switch data.(type) {
	//用户表
	case *SessionStoreData:
		cid := data.(*SessionStoreData).ClientID
		log.Info("cid : ", cid)
		_, err = op.Upsert(bson.M{"ClientID": cid}, data.(*SessionStoreData))
		if err != nil {
			log.Error(err.Error())
			return err
		}
	//添加群组
	case *TopicStoreData:
		TopicID := data.(*TopicStoreData).TopicID
		log.Info("TopicID : ", TopicID)
		_, err = op.Upsert(bson.M{"TopicID": TopicID}, data.(*TopicStoreData))
		if err != nil {
			log.Error(err.Error())
			return err
		}
	//P2P消息记录表
	case *P2PRecordMessageData:
		//消息记录储存
		FromID := data.(*P2PRecordMessageData).FromID
		log.Info("save p2p message : ", FromID)
		_, err = op.Upsert(bson.M{"From": FromID}, data.(*P2PRecordMessageData))
		if err != nil {
			log.Error(err.Error())
			return err
		}
	//Topic消息记录表
	case *TopicRecordMessageData:
		//消息记录储存
		FromID := data.(*TopicRecordMessageData).FromID
		log.Info("save topic message : ", FromID)
		_, err = op.Upsert(bson.M{"From": FromID}, data.(*TopicRecordMessageData))
		if err != nil {
			log.Error(err.Error())
			return err
		}
	//Mutual消息记录表
	case *MutualRecordMessageData:
		FromID := data.(*MutualRecordMessageData).FromID
		log.Info("save topic message : ", FromID)
		_, err = op.Upsert(bson.M{"From": FromID}, data.(*MutualRecordMessageData))
		if err != nil {
			log.Error(err.Error())
			return err
		}
	//KV表
	case *KVData:
		dType := data.(*KVData).Type
		key := data.(*KVData).Key
		_, err = op.Upsert(bson.M{"Type": dType, "Key": key}, data.(*KVData))
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}

	return err
}

//删除

func (self *MongoStore) Close() {
	self.session.Close()
}
