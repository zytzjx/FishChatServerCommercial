package mongo_store

import (
	"goProject/log"
	"gopkg.in/mgo.v2/bson"
)

//储存一些配置什么的
type KVData struct {
	Type  string   `bson:"Type"`
	Key   string   `bson:"Key"`
	Value []string `bson:"Value"`
}

//读取所有Router管理的所有MsgServer记录
func (self *MongoStore) ReadRouterManagerMsgServers(db string, c string, routers []string) []*KVData {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result []*KVData
	op.Find(bson.M{"Type": "routerManageMsgServer", "Key": bson.M{"$in": routers}}).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}
