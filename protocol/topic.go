
package protocol

import (
	"goProject/libnet"
)

type TopicMap   map[string]*Topic

type Topic struct {
	TopicName     string
	MsgAddr       string
	Channel       *libnet.Channel
	TA            *TopicAttribute
	ClientIDList  []string
}

func NewTopic(topicName string, msgAddr string, CreaterID string, CreaterSession *libnet.Session) *Topic {
	return &Topic {
		TopicName    : topicName,
		MsgAddr      : msgAddr,
		Channel      : new(libnet.Channel),
		TA           : NewTopicAttribute(CreaterID, CreaterSession),
		ClientIDList : make([]string, 0),
	}
}

//func (self *Topic)AddMember(m *redis_store.Member) {
//	self.TSD.MemberList = append(self.TSD.MemberList, m)
//}

type TopicAttribute struct {
	CreaterID          string
	CreaterSession     *libnet.Session
}

func NewTopicAttribute(CreaterID string, CreaterSession *libnet.Session) *TopicAttribute {
	return &TopicAttribute {
		CreaterID      : CreaterID,
		CreaterSession : CreaterSession,
	}
}