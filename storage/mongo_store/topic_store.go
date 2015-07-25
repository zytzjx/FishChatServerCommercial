

package mongo_store

type Member struct {
	ID   string
}

func NewMember(ID string) *Member {
	return &Member {
		ID : ID,
	}
}

type TopicStoreData struct {	
	TopicName      string      `bson:"TopicName"`
	CreaterID      string      `bson:"CreaterID"`
	MemberList     []*Member   `bson:"MemberList"`
	MsgServerAddr  string      `bson:"MsgServerAddr"`
}

func NewTopicStoreData(topicName string, createrID string, msgServerAddr string) *TopicStoreData {
	return &TopicStoreData{
		TopicName     : topicName,
		CreaterID     : createrID,
		MemberList    : make([]*Member, 0),
		MsgServerAddr : msgServerAddr,
	}
}
