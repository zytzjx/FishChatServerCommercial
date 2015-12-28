package mongo_store

import (
	"goProject/log"
	// "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//用户信息表
type SessionStoreData struct {
	ClientID      string   `bson:"ClientID"`
	ClientAddr    string   `bson:"ClientAddr"`
	MsgServerAddr string   `bson:"MsgServerAddr"`
	Friends       []string `bson:"Friends"`
	Alive         bool     `bson:"Alive"`
	Platform      string   `json:"Platform"`
}

//查询用户基本信息
type SessionStoreDataSimple struct {
	ClientID      string `bson:"ClientID"`
	ClientAddr    string `bson:"ClientAddr"`
	MsgServerAddr string `bson:"MsgServerAddr"`
	Alive         bool   `bson:"Alive"`
}

//查询用户基本信息[好友]
type SessionStoreDataFriends struct {
	ClientID string `bson:"ClientID"`
	Alive    bool   `bson:"Alive"`
}

func (self *MongoStore) UpdateSessionAlive(db string, c string, cid string, alive bool) error {
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	err = op.Update(bson.M{"ClientID": cid}, bson.M{"$set": bson.M{"Alive": alive}})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}

//通过CLientID查询信息
func (self *MongoStore) GetClientFromId(db string, c string, cid string) (*SessionStoreData, error) {
	log.Info("MongoStore GetClientFromId")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	var result *SessionStoreData

	err = op.Find(bson.M{"ClientID": cid}).One(&result)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return result, nil
}

//获取一组在线用户信息
func (self *MongoStore) GetOnlineClientsFromIds(db string, c string, ids []string) []*SessionStoreData {
	log.Info("MongoStore GetOnlineClientsFromIds")
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	var result []*SessionStoreData

	op.Find(bson.M{"ClientID": bson.M{"$in": ids}, "Alive": true}).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}

//获取一组用户信息
func (self *MongoStore) GetClientsFromIds(db string, c string, ids []string) []*SessionStoreData {
	log.Info("MongoStore GetClientsFromIds")
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	var result []*SessionStoreData

	op.Find(bson.M{"ClientID": bson.M{"$in": ids}}).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}

//获取一组好友信息
func (self *MongoStore) GetFriendsFromIds(db string, c string, ids []string) []*SessionStoreDataFriends {
	log.Info("MongoStore GetFriendsFromIds")
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	var result []*SessionStoreDataFriends

	op.Find(bson.M{"ClientID": bson.M{"$in": ids}}).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}

func (self *MongoStore) IsSessionAlive(db string, c string, cid string) (bool, error) {
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result *SessionStoreData

	err = op.Find(bson.M{"ClientID": cid}).One(&result)
	if err != nil {
		log.Error(err.Error())
		return false, err
	}

	return result.Alive, err
}
