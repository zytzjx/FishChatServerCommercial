package main

//router path
const (
	ROUTER_HOME          = "/"
	ROUTER_P2P_HISTORY   = "/history/v1/p2pHistory"
	ROUTER_TOPIC_HISTORY = "/history/v1/topicHistory"
	ROUTER_USER_REGISTER = "/user/v1/register"
)

const (
	ROUTER_VIEW_FRIEND = "/friend/v1/viewFriend"
	ROUTER_ADD_FRIEND  = "/friend/v1/addFriend"
)

//resp status
const (
	//Error
	RESP_STATUS_ERROR = "9999"
	//Success
	RESP_STATUS_SUCCESS = "0000"
	//Repeat registration
	RESP_STATUS_REPEAT_REGISTRATION = "9001"
)

//default settings
const (
	DEFAULT_GET_MSG_NUM = 100
)
