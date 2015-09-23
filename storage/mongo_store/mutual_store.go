package mongo_store

import (
	"goProject/log"
	"gopkg.in/mgo.v2/bson"
)

type MutualRecordMessageData struct {
	FromID string `bson:"FromID"` //来自用户ID
	ToID   string `bson:"ToID"`   //发送到某人ID
	Type   string `bson:"Type"`   //发送类型 addFriend,
	Time   int64  `bson:"Time"`   //时间
	UUID   string `bson:"UUID"`   //消息唯一标识符
	IsRead bool   `bson:"IsRead"` //是否已读
}

//读取单条未读消息记录
func (self *MongoStore) ReadMutualRecordMessageFromUuid(db string, c string, uuid string) *MutualRecordMessageData {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result *MutualRecordMessageData
	op.Find(bson.M{"UUID": uuid, "IsRead": false}).One(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}

//删除单条消息记录
func (self *MongoStore) RemoveMutualRecordMessageFromUuid(db string, c string, uuid string) error {
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	err = op.Remove(bson.M{"UUID": uuid})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}
