package mongo_store

import (
	"goProject/log"
	"gopkg.in/mgo.v2/bson"
)

// //查找好友列表
// func (self *MongoStore) GetFriendsFromIds(db string, c string, ids []string) []*SessionStoreDataBase {
// 	log.Info("MongoStore GetFriendsFromId")
// 	self.rwMutex.Lock()
// 	defer self.rwMutex.Unlock()

// 	op := self.session.DB(db).C(c)

// 	var result []*SessionStoreDataBase

// 	op.Find(bson.M{"ClientID": bson.M{"$in": ids}}).All(&result)
// 	defer func() {
// 		if err := recover(); err != nil {
// 			log.Error(err)
// 		}
// 	}()

// 	return result
// }

//修改好友列表
func (self *MongoStore) UpdateFriendsFromId(db string, c string, cid string, ids []string) error {
	log.Info("MongoStore UpdateFriendsFromId")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	op := self.session.DB(db).C(c)

	err = op.Update(bson.M{"ClientID": cid}, bson.M{"$set": bson.M{"Friends": ids}})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}
