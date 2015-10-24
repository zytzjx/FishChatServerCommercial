package mongo_store

const (
	DATA_BASE_NAME                   = "im"                    //数据库名
	CLIENT_INFO_COLLECTION           = "client_info"           //用户
	RECORD_P2P_MESSAGE_COLLECTION    = "p2p_record_message"    //消息记录
	TOPIC_INFO_COLLECTION            = "topic_info"            //群组
	RECORD_TOPIC_MESSAGE_COLLECTION  = "topic_record_message"  //群组消息记录
	RECORD_MUTUAL_MESSAGE_COLLECTION = "mutual_record_message" //用户交互消息记录
	KV_COLLECTION                    = "kvs"                   //kv配置数据
)
