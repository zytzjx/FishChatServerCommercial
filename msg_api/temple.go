package main

type EmptyTemple struct {
}

//基础返回
type BaseResultTemple struct {
	Status string      `json:"status"`
	Result interface{} `json:"result"`
}

//返回消息类型
type MsgsResultTemple struct {
	Total    int         `json:"total"`
	PageSize int         `json:"pageSize"`
	EndTime  int64       `json:"endTime"`
	Data     interface{} `json:"data"`
}

//p2p 消息返回格式
type P2PMsgTemple struct {
	MsgType  string `json:"msgType"`
	FromID   string `json:"fromId"`
	FriendId string `json:"friendId"`
	Content  string `json:"content"`
	Time     int64  `json:"time"`
	UUID     string `json:"uuid"`
}

//topic 消息返回格式
type TopicMsgTemple struct {
	MsgType string `json:"msgType"`
	FromID  string `json:"fromId"`
	TopicID string `json:"topicId"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
	UUID    string `json:"uuid"`
}

//friend
type FriendAliveResultTemple struct {
	FriendAlive []string `json:"friends_alive"`
	// FriendId  string
	// Alive     bool
}

type FriendAliveTemple struct {
	FriendId string `json:"friendId"`
	Alive    bool   `json:"alive"`
}
