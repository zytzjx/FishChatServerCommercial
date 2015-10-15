package protocol

//---------------------------------------------------------------------------
// General 通用协议
//---------------------------------------------------------------------------
const (
	//SEND_PING
	SEND_PING_CMD = "send_ping"
	//RESP_PONG_CMD
	RESP_PONG_CMD = "resp_pong"
)

//---------------------------------------------------------------------------
// Gateway
//---------------------------------------------------------------------------
const (
	REQ_MSG_SERVER_CMD = "req_msg_server"
	//SELECT_MSG_SERVER_FOR_CLIENT msg_server_ip
	SELECT_MSG_SERVER_FOR_CLIENT_CMD = "select_msg_server_for_client"
)
const (
	SEND_CLIENT_ID_CMD_ARGS_NUM = 1
)

//---------------------------------------------------------------------------
// Msg_server
//---------------------------------------------------------------------------
const (
	//SEND_CLIENT_ID CLIENT_ID
	SEND_CLIENT_ID_CMD = "send_client_id"
	//SEND_CLIENT_ID CLIENT_ID
	RESP_CLIENT_ID_CMD = "resp_client_id"

	SEND_CHANGE_MESSAGE_SERVER_CMD = "send_change_message_server"

	//SEND_LOGOUT_CMD
	SEND_LOGOUT_CMD = "send_logout"
	RESP_LOGOUT_CMD = "resp_logout"

	//SEND_CLIENT_ID_FOR_TOPIC ID
	SEND_CLIENT_ID_FOR_TOPIC_CMD = "send_client_id_for_topic"
	//SUBSCRIBE_CHANNEL channelName
	SUBSCRIBE_CHANNEL_CMD = "subscribe_channel"
	//SEND_MESSAGE_P2P send2msg send2ID toID
	SEND_MESSAGE_P2P_CMD = "send_message_p2p"
	//RESP_MESSAGE_P2P  msg fromID time uuid
	RESP_MESSAGE_P2P_CMD = "resp_message_p2p"

	SEND_NOTIFY_P2P_CMD = "send_notify_p2p"
	RESP_NOTIFY_P2P_CMD = "resp_notify_p2p"

	//RECEIVE_MESSAGE_P2P_CMD
	RECEIVE_MESSAGE_P2P_CMD = "receive_message_p2p"
	RECEIVE_NOTIFY_P2P_CMD  = "receive_notify_p2p"

	//CREATE_TOPIC TOPIC_NAME
	SEND_CREATE_TOPIC_CMD = "send_create_topic"
	//RESP TOPIC_NAME
	RESP_CREATE_TOPIC_CMD = "resp_create_topic"

	//JOIN_TOPIC TOPIC_NAME
	SEND_JOIN_TOPIC_CMD = "send_join_topic"
	//RESP TOPIC_NAME
	RESP_JOIN_TOPIC_CMD = "resp_join_topic"

	//SEND_LEAVE_TOPIC_CMD TOPIC_NAME
	SEND_LEAVE_TOPIC_CMD = "send_leave_topic"
	//RESP TOPIC_NAME
	RESP_LEAVE_TOPIC_CMD = "resp_leave_topic"

	//SEND_LIST_TOPIC_CMD
	SEND_LIST_TOPIC_CMD = "send_list_topic"
	//RESP_LIST_TOPIC_CMD Topiclist
	RESP_LIST_TOPIC_CMD = "resp_list_topic"

	//SEND_TOPIC_MEMBERS_LIST_CMD TOPIC_NAME
	SEND_TOPIC_MEMBERS_LIST_CMD = "send_topic_members_list"
	//RESP_TOPIC_MEMBERS_LIST_CMD TOPIC_NAME MembersList
	RESP_TOPIC_MEMBERS_LIST_CMD = "resp_topic_members_list"

	SEND_LOCATE_TOPIC_MSG_ADDR_CMD = "send_locate_topic_msg_addr"
	//SEND_MESSAGE_TOPIC_CMD send2msg topicId fromId

	SEND_MESSAGE_TOPIC_CMD = "send_message_topic"
	//RESP_MESSAGE_TOPIC_CMD
	RESP_MESSAGE_TOPIC_CMD = "resp_message_topic"

	SEND_NOTIFY_TOPIC_CMD = "send_notify_topic"
	RESP_NOTIFY_TOPIC_CMD = "resp_notify_topic"

	//RECEIVE_MESSAGE_TOPIC_CMD send2Msg topicId fromId time uuid
	RECEIVE_MESSAGE_TOPIC_CMD = "receive_message_topic"
	RECEIVE_NOTIFY_TOPIC_CMD  = "receive_notify_topic_cmd"

	//SEND_VIEW_FRIENDS_CMD FRIEND_ID
	SEND_VIEW_FRIENDS_CMD = "send_view_friends"
	//SEND_VIEW_FRIENDS_CMD
	RESP_VIEW_FRIENDS_CMD = "resp_view_friends"

	//SEND_ADD_FRIEND_CMD FRIEND_ID
	SEND_ADD_FRIEND_CMD = "send_add_friend"
	//SEND_ADD_FRIEND_CMD
	RESP_ADD_FRIEND_CMD = "resp_add_friend"

	//SEND_DEL_FRIEND_CMD FRIEND_ID
	SEND_DEL_FRIEND_CMD = "send_del_friend"
	//SEND_DEL_FRIEND_CMD
	RESP_DEL_FRIEND_CMD = "resp_del_friend"

	// SEND_ASK_CMD type:add_friend,add_topic,invite_topic target
	SEND_ASK_CMD                   = "send_ask"
	SEND_ASK_CMD_TYPE_ADD_FRIEND   = "add_friend"
	SEND_ASK_CMD_TYPE_ADD_TOPIC    = "add_topic"
	SEND_ASK_CMD_TYPE_INVITE_TOPIC = "invite_topic"
	// RESP_ASK_CMD
	RESP_ASK_CMD = "resp_ask"

	// RECEIVE_ASK_CMD
	RECEIVE_ASK_CMD = "receive_ask"

	//SEND_REACT_CMD type:add_friend,add_topic,invite_topic target
	SEND_REACT_CMD                   = "send_react"
	SEND_REACT_CMD_TYPE_ADD_FRIEND   = "add_friend"
	SEND_REACT_CMD_TYPE_ADD_TOPIC    = "add_topic"
	SEND_REACT_CMD_TYPE_INVITE_TOPIC = "invite_topic"
	SEND_REACT_CMD_AGREE             = "agree"
	SEND_REACT_CMD_DISAGREE          = "disagree"
	//RESP_REACT_CMD
	RESP_REACT_CMD = "resp_react"

	//SEND_MONITOR_INFO_CMD time, infoJson
	SEND_MONITOR_INFO_CMD = "send_monitor_info"
)
const (
	SEND_MESSAGE_P2P_CMD_ARGS_NUM           = 2
	P2P_ACK_CMD_ARGS_NUM                    = 1
	SEND_CHANGE_MESSAGE_SERVER_CMD_ARGS_NUM = 2
	TOPIC_ACK_CMD_ARGS_NUM                  = 1
	SEND_CREATE_TOPIC_CMD_ARGS_NUM          = 1
	SEND_JOIN_TOPIC_CMD_ARGS_NUM            = 1
	SEND_LEAVE_TOPIC_CMD_ARGS_NUM           = 1
	SEND_TOPIC_MEMBERS_LIST_CMD_ARGS_NUM    = 1
	SEND_MESSAGE_TOPIC_CMD_ARGS_NUM         = 2
	SEND_ADD_FRIEND_CMD_ARGS_NUM            = 1
	SEND_DEL_FRIEND_CMD_ARGS_NUM            = 1
	SEND_ASK_CMD_ARGS_NUM                   = 2
	SEND_REACT_CMD_ARGS_NUM                 = 2
	ASK_ACK_CMD_ARGS_NUM                    = 2
	REACT_ACK_CMD_ARGS_NUM                  = 2
	SEND_MONITOR_INFO_CMD_ARGS_NUM          = 2
)
const (
	//P2P_ACK uuid
	P2P_ACK_CMD      = "p2p_ack"
	P2P_ACK_FAILURES = 3
	P2P_ACK_TIMEOUT  = 3

	//TOPIC_ACK uuid
	TOPIC_ACK_CMD      = "topic_ack"
	TOPIC_ACK_FAILURES = 3
	TOPIC_ACK_TIMEOUT  = 3

	//ASK_ACK_CMD uuid
	ASK_ACK_CMD      = "ask_ack"
	ASK_ACK_FAILURES = 3
	ASK_ACK_TIMEOUT  = 3
)

//---------------------------------------------------------------------------
// Router
//---------------------------------------------------------------------------
const (
	//ROUTE_MSG_CMD MsgServer BSONCMD
	ROUTE_MSG_CMD = "route_msg"

	//CHANGE_MESSAGE_SERVER_CMD cid (由router转发,如果用户在另外一台message_server登陆,就发送断开请求到另外一台服务器)
	ROUTE_CHANGE_MESSAGE_SERVER_CMD = "route_change_message_server"

	//ROUTE_MESSAGE_P2P_CMD  msg fromID time uuid
	ROUTE_MESSAGE_P2P_CMD = "route_message_p2p"
	ROUTE_NOTIFY_P2P_CMD  = "route_notify_p2p"

	//ROUTE_MESSAGE_TOPIC_CMD
	ROUTE_MESSAGE_TOPIC_CMD = "route_message_topic"
	ROUTE_NOTIFY_TOPIC_CMD  = "route_notify_topic"

	//ROUTE_ASK_CMD
	ROUTE_ASK_CMD = "route_ask"
)
const (
	ROUTE_MSG_CMD_ARGS_NUM                   = 2
	ROUTE_MESSAGE_P2P_CMD_ARGS_NUM           = 5
	ROUTE_CHANGE_MESSAGE_SERVER_CMD_ARGS_NUM = 1
	ROUTE_MESSAGE_TOPIC_CMD_ARGS_NUM         = 6
	ROUTE_ASK_CMD_ARGS_NUM                   = 5
)

//---------------------------------------------------------------------------
// monitor
//---------------------------------------------------------------------------
const (
	TYPE_GATEWAY_SERVER    = "gateway"
	TYPE_ROUTER_SERVER     = "router"
	TYPE_MSG_SERVER_SERVER = "msg_server"
)

// const (
// 	CACHE_SESSION_CMD = "cache_session"
// 	CACHE_TOPIC_CMD   = "cache_topic"
// )

// const (
// 	STORE_SESSION_CMD = "store_session"
// 	STORE_TOPIC_CMD   = "store_topic"
// )

// const (
// 	PING = "ping"
// )

//---------------------------------------------------------------------------
// 数据协议接口
//---------------------------------------------------------------------------
type Cmd interface {
	GetCmdName() string
	ChangeCmdName(newName string)
	GetArgs() []string
	AddArg(arg string)
	ParseCmd(msglist []string)
	GetAnyData() interface{}
	GetReport() interface{}
}

//---------------------------------------------------------------------------
// 服务器发送消息给客户端数据格式
// {"cmd":"req_msg_server","obj":[]}
//---------------------------------------------------------------------------
type CmdSimple struct {
	CmdName string      `json:"cmd"`
	Args    []string    `json:"obj"`
	Repo    interface{} `json:"repo"`
}

func NewCmdSimple(cmdName string) *CmdSimple {
	return &CmdSimple{
		CmdName: cmdName,
		Args:    make([]string, 0),
	}
}

func (self *CmdSimple) GetCmdName() string {
	return self.CmdName
}

func (self *CmdSimple) ChangeCmdName(newName string) {
	self.CmdName = newName
}

func (self *CmdSimple) GetArgs() []string {
	return self.Args
}

func (self *CmdSimple) AddArg(arg string) {
	self.Args = append(self.Args, arg)
}

func (self *CmdSimple) ParseCmd(msglist []string) {
	self.CmdName = msglist[1]
	self.Args = msglist[2:]
}

func (self *CmdSimple) GetAnyData() interface{} {
	return nil
}

func (self *CmdSimple) GetReport() interface{} {
	return self.Repo
}

//---------------------------------------------------------------------------
// 服务器发送消息给客户端数据格式
// {"cmd":"select_msg_server_for_client","ok":true,"msg":"",
// "obj":["127.0.0.1:19000"]}
//---------------------------------------------------------------------------
type CmdResponse struct {
	CmdName string      `json:"cmd"`
	Ok      bool        `json:"ok"`
	Message string      `json:"msg"`
	Args    []string    `json:"obj"`
	Repo    interface{} `json:"repo"`
}

func NewCmdResponse(cmdName string) *CmdResponse {
	return &CmdResponse{
		CmdName: cmdName,
		Ok:      true,
		Args:    make([]string, 0),
	}
}

func (self *CmdResponse) GetCmdName() string {
	return self.CmdName
}

func (self *CmdResponse) ChangeCmdName(newName string) {
	self.CmdName = newName
}

func (self *CmdResponse) GetArgs() []string {
	return self.Args
}

func (self *CmdResponse) AddArg(arg string) {
	self.Args = append(self.Args, arg)
}

func (self *CmdResponse) GetAnyData() interface{} {
	return nil
}

//---------------------------------------------------------------------------
// Monitor服务器收集信息数据格式
// {"type":"msg_server","time":1426809438,"data":{"SessionNum":56}}
//---------------------------------------------------------------------------

type CmdMonitor struct {
	ServerType string `json:"type"`
	Time       int64  `json:"time"`
	Data       string `json:"data"`
}

func NewCmdMonitor(serverType string, time int64, data string) *CmdMonitor {
	return &CmdMonitor{
		ServerType: serverType,
		Time:       time,
		Data:       data,
	}
}

type MsgServerMonitorData struct {
	SessionNum uint64 `json:"session_num"`
}
