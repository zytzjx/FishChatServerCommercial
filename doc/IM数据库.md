
# GOLANG IM 数据库结构

> <small>nprog | <ingram@60.com> | 2015/9/23 | version: 0.1</small>

###目录

1. [client_info 用户信息表](#client_info)
2. [topic_info 群组信息表](#topic_info)
3. [kvs 键值对表(Router对应Msg_server的记录)](#kvs)
4. [p2p_record_message P2P消息记录表](#p2p_record_message)
5. [topic_record_message Topic消息记录表](#topic_record_message)
6. [mutual_record_message 交互信息记录表](#mutual_record_message)


###文档


<a name="client_info"></a>
> client_info 用户信息表 | IM用户基本信息和在线状态

表结构

    ClientID      string   `bson:"ClientID"`        //用户登陆ID
    ClientAddr    string   `bson:"ClientAddr"`      //用户ip:port
    MsgServerAddr string   `bson:"MsgServerAddr"`   //用户连接msg_server ip:port
    Friends       []string `bson:"Friends"`         //好友列表
    Alive         bool     `bson:"Alive"`           //是否在线

示例数据
```
{
    "_id" : ObjectId("55ead45caabd3f08f5a921ca"),
    "ClientID" : "aa",
    "ClientAddr" : "192.168.60.101:58604",
    "MsgServerAddr" : "127.0.0.1:19000",
    "Friends" : [
        "ee",
        "bb",
        "cc",
        "ll"
    ],
    "Alive" : true
}
```
[TOP](#)
<hr>

<a name="topic_info"></a>
> topic_info 群组信息表 | IM群组基本信息

表结构

    TopicID       string        //群组ID
    FounderID     string        //创建者
    ClientsID     []string      //成员[u1, u2, u3]


示例数据
```
{
    "_id" : ObjectId("55e869d5aabd3f08f5a921bb"),
    "TopicID" : "aa",
    "FounderID" : "cc",
    "ClientsID" : [
        "aa",
        "bb",
        "cc"
    ],
}

```
[TOP](#)
<hr>

<a name="kvs"></a>
> kvs 键值对表 | 目前只储存Router对应Msg_server的记录

表结构

    Type  string   //记录类型
    Key   string   //键
    Value []string //值

示例数据
```
{
    "_id" : ObjectId("55f49e6daabd3f08f5a92200"),
    "Type" : "routerManageMsgServer",
    "Key" : "127.0.0.1:20000",
    "Value" : [
        "127.0.0.1:19000",
        "127.0.0.1:19002"
    ]
}

```
[TOP](#)
<hr>


<a name="p2p_record_message"></a>
> p2p_record_message P2P消息记录表 | 记录P2P离线消息

表结构

    FromID      string      //来自用户ID
    ToID        string      //发送到某人ID
    Content     string      //消息内容
    Time        int64       //时间
    UUID        string      //消息唯一标识符
    IsRead      bool        //是否已读

示例数据
```
{
        "_id" : ObjectId("55d39b65aabd3f08f5a9219e"),
        "FromID" : "aa",
        "ToID" : "bb",
        "Content" : "Hello world",
        "Time" : NumberLong(1439931237),
        "UUID" : "d430f0bd-2d24-4bcb-98ef-cb1c0e5da69c",
        "IsRead" : true
}
```
[TOP](#)
<hr>

<a name="topic_record_message"></a>
> topic_record_message Topic消息记录表 | 记录Topic离线消息

表结构

    FromID      string      //来自用户ID
    ToID        string      //发送到Topic ID
    Content     string      //消息内容
    Time        int64       //时间
    UUID        string      //消息唯一标识符
    IsRead      []string    //是否已读 储存格式 [u1, u2, u3]

示例数据
```
{
        "_id" : ObjectId("55eb1412aabd3f08f5a921ce"),
        "FromID" : "cc",
        "ToID" : "aa",
        "Content" : "23571q92435",
        "Time" : NumberLong(1441469458),
        "UUID" : "a508e120-0bb6-49ee-959a-b07595acfe91",
        "IsRead" : [
                "aa",
                "cc",
                "bb"
        ]
}
```
[TOP](#)
<hr>

<a name="mutual_record_message"></a>
> mutual_record_message 交互信息记录表 | 记录Mutual离线消息

表结构

    FromID      string      //来自用户ID
    ToID        string      //发送到某人ID
    Type        string      //发送类型 addFriend,
    Time        int64       //时间
    UUID        string      //消息唯一标识符
    IsRead      bool        //是否已读

示例数据
```
{
        "_id" : ObjectId("55f1fb8caabd3f08f5a921eb"),
        "FromID" : "aa",
        "ToID" : "bb",
        "Type" : "add_friend",
        "Time" : NumberLong(1441921932),
        "UUID" : "e0b64e21-123a-4031-9fdc-647fe2d36142",
        "IsRead" : false
}
```
[TOP](#)
<hr>