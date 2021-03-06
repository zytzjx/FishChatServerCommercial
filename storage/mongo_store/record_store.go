package mongo_store

import (
	"goProject/log"
	// "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//消息记录储存
type P2PRecordMessageData struct {
	MsgType string `bson:"MsgType"` //消息类型
	FromID  string `bson:"FromID"`  //来自用户ID
	ToID    string `bson:"ToID"`    //发送到某人ID
	Content string `bson:"Content"` //消息内容
	Time    int64  `bson:"Time"`    //时间
	UUID    string `bson:"UUID"`    //消息唯一标识符
	IsRead  bool   `bson:"IsRead"`  //是否已读
}

//群组消息储存
type TopicRecordMessageData struct {
	MsgType string   `bson:"MsgType"` //消息类型
	FromID  string   `bson:"FromID"`  //来自用户ID
	ToID    string   `bson:"ToID"`    //发送到Topic ID
	Content string   `bson:"Content"` //消息内容
	Time    int64    `bson:"Time"`    //时间
	UUID    string   `bson:"UUID"`    //消息唯一标识符
	IsRead  []string `bson:"IsRead"`  //是否已读 储存格式 [u1, u2, u3]
}

//读取未读消息记录
func (self *MongoStore) ReadP2PRecordMessage(db string, c string, cid string) ([]*P2PRecordMessageData, error) {
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result []*P2PRecordMessageData
	err = op.Find(bson.M{"ToID": cid, "IsRead": false}).All(&result)

	if err != nil {
		log.Error(err.Error())
		return result, err
	}

	return result, err
}

//读取一个用户所有未读消息数量
func (self *MongoStore) ReadP2PRecordNumber(db string, c string, cid string) int {
	var (
		err error
		num int
	)
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	num, err = op.Find(bson.M{"ToID": cid, "IsRead": false}).Count()

	if err != nil {
		log.Error(err.Error())
		return 0
	}

	return num
}

//读取单条未读消息记录
func (self *MongoStore) ReadP2PRecordMessageFromUuid(db string, c string, uuid string) *P2PRecordMessageData {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result *P2PRecordMessageData
	op.Find(bson.M{"UUID": uuid, "IsRead": false}).One(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}

//根据UUID标记群组未读消息记录
func (self *MongoStore) MarkP2PRecordMessageFromUuid(db string, c string, uuid string) error {
	log.Info("::Set record message to readed")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	_, err = op.UpdateAll(bson.M{"UUID": uuid}, bson.M{"$set": bson.M{"IsRead": true}})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}

//读取群组未读消息记录
func (self *MongoStore) ReadTopicRecordMessage(db string, c string, cid string, topicIds []string) []*TopicRecordMessageData {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result []*TopicRecordMessageData
	//查找IsRead中不包含ClientID的记录

	op.Find(bson.M{"ToID": bson.M{"$in": topicIds}, "IsRead": bson.M{"$ne": cid}}).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	return result
}

//根据UUID读取群组未读信息
func (self *MongoStore) ReadTopicRecordMessageFromUuid(db string, c string, uuid string) *TopicRecordMessageData {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result *TopicRecordMessageData

	op.Find(bson.M{"UUID": uuid}).One(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	return result
}

//根据传入时间往前取N条P2P历史纪录
func (self *MongoStore) ReadP2PHistoryFromEndTime(db string, c string, FromID string, ToID string, endTime int64, n int) []*P2PRecordMessageData {
	log.Info("::ReadP2PHistoryFromEndTime")
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result []*P2PRecordMessageData

	op.Find(bson.M{"FromID": bson.M{"$in": []string{FromID, ToID}}, "ToID": bson.M{"$in": []string{FromID, ToID}}, "Time": bson.M{"$lte": endTime}}).Sort("-Time").Limit(n).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	return result
}

//根据传入时间往前取N条群组历史纪录
func (self *MongoStore) ReadTopicHistoryFromEndTime(db string, c string, topicName string, endTime int64, n int) []*TopicRecordMessageData {
	log.Info("::ReadHistoryFromEndTime")
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	var result []*TopicRecordMessageData

	op.Find(bson.M{"ToID": topicName, "Time": bson.M{"$lte": endTime}}).Sort("-Time").Limit(n).All(&result)
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	return result
}

//根据UUID标记群组未读消息记录
func (self *MongoStore) MarkTopicRecordMessageFromUuid(db string, c string, uuid string, readStr []string) error {
	log.Info("::Set record message to readed")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	err = op.Update(bson.M{"UUID": uuid}, bson.M{"$set": bson.M{"IsRead": readStr}})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return err
}

//根据用户和时间标记之前的所有群组信息为已读
func (self *MongoStore) MarkTopicRecordMessageFromUserAndTime(db string, c string, user string, timeNow int64, topicName string) error {
	log.Info("::MarkTopicRecordMessageFromUserAndTime")
	var err error
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	op := self.session.DB(db).C(c)

	_, err = op.UpdateAll(bson.M{"Time": bson.M{"$lt": timeNow}, "ToID": topicName}, bson.M{"$addToSet": bson.M{"IsRead": user}})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return err
}
