
# GOLANG IM API 开发文档

> <small>nprog | <ingram@60.com> | 2015/10/13 | version: 0.2.2</small>


###目录

1. [GI-001-G 接口描述：请求分配msg_server](#GI-001-G)
2. [GI-002-M 请求登录msg_server](#GI-002-M)
3. [GI-003-M 退出msg_server](#GI-003-M)
4. [GI-004-M P2P发送消息](#GI-004-M)
5. [GI-005-M P2P接收消息](#GI-005-M)
6. [GI-006-M Client发送PING给Msg_server](#GI-006-M)
7. [GI-007-M 获取用户群组列表](#GI-007-M)
8. [GI-008-M Topic新建指令](#GI-008-M)
9. [GI-009-M Topic加入](#GI-009-M)
10. [GI-010-M Topic退出](#GI-010-M)
11. [GI-011-M Topic获取用户列表](#GI-011-M)
12. [GI-012-M Topic发送群消息](#GI-012-M)
13. [GI-013-M Topic接收群消息](#GI-013-M)
14. [GI-014-M 请求好友列表](#GI-014-M)
15. [GI-015-M 添加好友请求(单向)](#GI-015-M)
16. [GI-016-M 删除好友请求](#GI-016-M)
17. [GI-017-M 发送交互请求统一处理接口](#GI-017-M)
18. [GI-018-M 接收交互请求统一处理接口](#GI-018-M)
19. [GI-019-M 回复交互请求操作接口](#GI-019-M)
20. [GI-020-M P2P发送通知](#GI-020-M)
21. [GI-021-M P2P接收消息](#GI-021-M)
21. [GI-022-M Topic发送群通知](#GI-022-M)
22. [GI-023-M Topic接收群通知](#GI-023-M)

###更新日志
> 
>  * 2015/10/13 | version: 0.2.2
>  <small>加入一组(GI-020-M~GI-023-M)通知类命令，用法与普通消息类似。用于发送系统通知，如添加删除好友成功信息等。</small>
> 


###文档

<a name="GI-001-G"></a>
> 序号:GI-001-G | 接口描述：请求分配msg_server | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to gateway
数据标示符
    req_msg_server

数据格式
    {cmd [] repo}

数据样例    
    {"cmd":"req_msg_server", "obj":[], "repo":null}
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Gateway to client
数据标示符
    select_msg_server_for_client
数据格式
    {cmd ok msg [msg_server主机地址] repo}
数据样例
    {"cmd":"select_msg_server_for_client","ok":true,"msg":"","obj":["127.0.0.1:19000"], "repo":null}
备注
    分配到的msg_server的主机为127.0.0.1:19000
```
[TOP](#)

---

<a name="GI-002-M"></a>
> 序号:GI-002-M | 接口描述：请求登录msg_server | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_client_id

数据格式
    {cmd [ClientID] repo}

数据样例    
    {"cmd":"send_client_id","obj":["aa"], "repo":null}  //发送我的ID aa给服务器
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_client_id
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_client_id","ok":true,"msg":"","obj":[], "repo":null}//成功
    {"cmd":"resp_client_id","ok":false,"msg":"The user is logged","obj":[], repo:null}//失败
备注
    如果相同账号重复登陆，服务器会自动断开之前登陆的账号
```
[TOP](#)

---

<a name="GI-003-M"></a>
> 序号:GI-003-M | 接口描述：退出msg_server | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_logout

数据格式
    {cmd [] repo}

数据样例    
    {"cmd":"send_logout","obj":[], "repo":null}
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_logout
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_logout","ok":true,"msg":"","obj":[], "repo":null}//成功
备注
    服务器会在最后返回信息时断开连接
```
[TOP](#)

---

<a name="GI-004-M"></a>
> 序号:GI-004-M | 接口描述：P2P发送消息 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_message_p2p

数据格式
    {cmd [消息内容 目标ID] repo}

数据样例    
    {"cmd":"send_message_p2p","obj":["hello","alex"], "repo":null} //发消息给alex，内容为hello
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_message_p2p
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_message_p2p","ok":true,"msg":"","obj":[], "repo":null}//成功
    {"cmd":"resp_message_p2p","ok":false,"msg":"Not exists client","obj":[], repo:null}//失败
备注
```
[TOP](#)

---

<a name="GI-005-M"></a>
> 序号:GI-005-M | 接口描述：P2P接收消息 | 传输协议: TCP

```
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    receive_message_p2p

数据格式
    {cmd ok msg [消息内容 来自ID 发送时间 UUID] repo}

数据样例    
    {"cmd":"receive_message_p2p","ok":true,"msg":"","obj":["hello","bb","1442469179","7b4e412f-36e1-4fe7-8dfd-4e0899e79768"]
, "repo":null}
备注
    收到来自bb的消息，内容为hello
    
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    p2p_ack
数据格式
    {cmd ok msg [UUID] repo}
数据样例
    {"cmd":"p2p_ack","obj":["7b4e412f-36e1-4fe7-8dfd-4e0899e79768"], "repo":null}
备注
    确认收到该信息uuid为 7b4e412f-36e1-4fe7-8dfd-4e0899e79768 的消息
    接收P2P消息需要返回ACK信息，否则服务器将会在3秒后自动重发该消息给客户端
```
[TOP](#)

---

<a name="GI-006-M"></a>
> 序号:GI-006-M | 接口描述：Client发送PING给Msg_server | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_ping

数据格式
    {cmd [] repo}

数据样例    
    {"cmd":"send_ping","obj":[], "repo":null}
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_pong
数据格式
    {cmd [] repo}
数据样例
    {"cmd":"resp_pong","obj":[], "repo":null}
备注
```
[TOP](#)

---

<a name="GI-007-M"></a>
> 序号:GI-007-M | 接口描述：获取用户群组列表 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_list_topic

数据格式
    {cmd [] repo}

数据样例    
    {"cmd":"send_list_topic","obj":[], "repo":null}
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_list_topic
数据格式
    {cmd ok msg [Topics] repo}
数据样例
    {"cmd":"resp_list_topic","ok":true,"msg":"","obj":["[\"aa"\, \"bb\"]"], "repo":null}//成功
备注
```
[TOP](#)

---

<a name="GI-008-M"></a>
> 序号:GI-008-M | 接口描述：Topic 新建指令 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_create_topic

数据格式
    {cmd [TopicID] repo}

数据样例    
    {"cmd":"send_create_topic","obj":["t1"], "repo":null}
备注
    新建一个名字为t1的topic
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_create_topic
数据格式
    {cmd ok msg [TopicID] repo}
数据样例
    {"cmd":"resp_create_topic","ok":true,"msg":"","obj":["aa"], "repo":null}//成功
    {"cmd":"resp_create_topic","ok":false,"msg":"Create topic failure.","obj":[], "repo":null}//失败
备注
```
[TOP](#)

---

<a name="GI-009-M"></a>
> 序号:GI-009-M | 接口描述：Topic 用户加入 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_join_topic

数据格式
    {cmd [TopicID] repo}

数据样例    
    {"cmd":"send_join_topic","obj":["t1"], "repo":null}
备注
    加入一个名字为t1的topic
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_join_topic
数据格式
    {cmd ok msg [TopicID] repo}
数据样例
    {"cmd":"resp_join_topic","ok":true,"msg":"","obj":["aa"], "repo":null}//成功
    {"cmd":"resp_join_topic","ok":false,"msg":"Exceed the maximum number","obj":[], "repo":null}//失败
备注
```
[TOP](#)

---

<a name="GI-010-M"></a>
> 序号:GI-010-M | 接口描述：Topic 退出 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_leave_topic

数据格式
    {cmd [TopicID] repo}

数据样例    
    {"cmd":"send_leave_topic","obj":["t1"], "repo":null}
备注
    退出一个名字为t1的topic
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_leave_topic
数据格式
    {cmd ok msg [TopicID] repo}
数据样例
    {"cmd":"resp_leave_topic","ok":true,"msg":"","obj":["aa"], "repo":null}//成功
    {"cmd":"resp_leave_topic","ok":false,"msg":"Not exists topic name","obj":[], "repo":null}//失败
备注
```
[TOP](#)

---

<a name="GI-011-M"></a>
> 序号:GI-011-M | 接口描述：Topic 获取用户列表 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_topic_members_list

数据格式
    {cmd [TopicID] repo}

数据样例    
    {"cmd":"send_topic_members_list","obj":["t1"], "repo":null}
备注
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_topic_members_list
数据格式
    {cmd ok msg [TopicID] repo}
数据样例
    {"cmd":"resp_topic_members_list","ok":true,"msg":"","obj":["t","[{\"Client\":\"aa\",\"Alive\":true},{\"Client\":\"bb\",\"Alive\":false},{\"Client\":\"cc\",\"Alive\":true}]"],"repo":null} //成功
    {"cmd":"resp_topic_members_list","ok":false,"msg":"Not exists topic name","obj":[],"repo":null}//失败
备注
```
[TOP](#)

---

<a name="GI-012-M"></a>
> 序号:GI-012-M | 接口描述：Topic 发送群消息 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_message_topic

数据格式
    {cmd [Message TopicID] repo}

数据样例    
    {"cmd":"send_message_topic","obj":["hello", "t"], "repo":null}
备注
    发出消息到name为t的topic
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_message_topic
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_message_topic","ok":true,"msg":"","obj":[], "repo":null}//成功
    {"cmd":"resp_message_topic","ok":false,"msg":"You were not in this topic.","obj":[], repo:null}//失败
备注
```
[TOP](#)

---

<a name="GI-013-M"></a>
> 序号:GI-013-M | 接口描述：Topic 接收群消息 | 传输协议: TCP

```
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    receive_message_topic

数据格式
    {cmd ok msg [Message topicId fromId Time UUID] repo}

数据样例    
    {"cmd":"receive_message_topic","ok":true, "msg":"","obj":["hello", "t", "aa","1442469179","7b4e412f-36e1-4fe7-8dfd-4e0899e79768"], "repo":null}
备注
    接收到来自的Topic t的信息，内容为 hello， 发送人为aa
    
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    topic_ack
数据格式
    {cmd [UUID] repo}
数据样例
    {"cmd":"topic_ack","obj":["7b4e412f-36e1-4fe7-8dfd-4e0899e79768"]}
备注
```
[TOP](#)

---

<a name="GI-014-M"></a>
> 序号:GI-014-M | 接口描述：请求好友列表 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_view_friends

数据格式
    {cmd [FriendID] repo}

数据样例    
    {"cmd":"send_view_friends", "obj":[], "repo":null}
备注
    删除好友bb
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_view_friends
数据格式
    {cmd ok msg [list] repo}
数据样例
    {"cmd":"resp_view_friends","ok":true,"msg":"","obj":["[{\"client\":\"aa\",\"alive\":true},{\"client\":\"bb\",\"alive\":false}]"], "repo":null}
备注
```
[TOP](#)

---

<a name="GI-015-M"></a>
> 序号:GI-015-M | 接口描述：添加好友请求(单向) | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_add_friend

数据格式
    {cmd [FriendID] repo}

数据样例    
    {"cmd":"send_add_friend","obj":["bb"], "repo":null}
备注
    发送添加bb为好友的请求
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_add_friend
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_add_friend","ok":true,"msg":"","obj":[], "repo":null}//成功
    {"cmd":"resp_add_friend","ok":false,"msg":"The id is already your friend.","obj":[], repo:null}//失败
备注
```
[TOP](#)

---

<a name="GI-016-M"></a>
> 序号:GI-016-M | 接口描述：删除好友请求 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_del_friend

数据格式
    {cmd [FriendID] repo}

数据样例    
    {"cmd":"send_del_friend","obj":["bb"], "repo":null}
备注
    删除好友bb
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_del_friend
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_del_friend","ok":true,"msg":"","obj":[], "repo":null}//成功
    {"cmd":"resp_del_friend","ok":false,"msg":"You have no this friend.","obj":[], repo:null}//失败
备注
```
[TOP](#)

---

<a name="GI-017-M"></a>
> 序号:GI-017-M | 接口描述：发送交互请求统一处理接口 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_ask

数据格式
    {cmd [消息类型 目标ID] repo}

数据样例
    发送添加好友请求:
        {"cmd":"send_ask","obj":["add_friend","bb"], "repo":null}
备注
    目前消息类型只支持add_friend
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_ask
数据格式
    {cmd ok msg [] repo}
数据样例
    {"cmd":"resp_ask","ok":true,"msg":"","obj":[], "repo":null}//成功
备注
```
[TOP](#)

---

<a name="GI-018-M"></a>
> 序号:GI-018-M | 接口描述：接收交互请求统一处理接口 | 传输协议: TCP

```
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    receive_ask

数据格式
    {cmd ok msg [消息类型 来自ID 消息时间 UUID] repo}

数据样例
    接收添加好友请求:
        {"cmd":"receive_ask","ok":true,"msg":"","obj":["add_friend","aa","1442469179","9042388a-2519-4291-abc4-8cdbb0076456"], "repo":null}
备注
    目前消息类型只支持add_friend
```
[TOP](#)

---

<a name="GI-019-M"></a>
> 序号:GI-019-M | 接口描述：回复交互请求操作接口 | 传输协议: TCP

```
>>>>>>>>>>>>>>>>>>>>>>>>>Client to msg_server
数据标示符
    send_react

数据格式
    {cmd [处理方式 UUID] repo}

数据样例
    同意请求:
        {"cmd":"send_react","obj":["agree","9042388a-2519-4291-abc4-8cdbb0076456"], "repo":null}
    拒绝请求:
        {"cmd":"send_react","obj":["disagree","9042388a-2519-4291-abc4-8cdbb0076456"], "repo":null}
备注
    同意操作agree, 拒绝操作disagree
    如果是同意操作，会自动发一条信息给对方
    
<<<<<<<<<<<<<<<<<<<<<<<<<Msg_server to client
数据标示符
    resp_react
数据格式
    {cmd ok msg [FriendID] repo}
数据样例
    {"cmd":"resp_react","ok":true,"msg":"","obj":["bb"], "repo":null} //成功
备注
```
[TOP](#)

---

<a name="GI-020-M"></a>
> 序号:GI-020-M | 接口描述：P2P发送通知 | 传输协议: TCP

<small>
发送数据标识符：send_notify_p2p
回执数据标识符：resp_notify_p2p
</small>
该命令为P2P发送消息的别名
用法参见：[GI-004-M P2P发送消息](#GI-004-M)

[TOP](#)

---

<a name="GI-021-M"></a>
> 序号:GI-021-M | 接口描述：P2P接收通知 | 传输协议: TCP

<small>
发送数据标识符：receive_notify_p2p
回执数据标识符：p2p_ack
</small>
该命令为P2P接收消息的别名
用法参见：[GI-005-M P2P接收消息](#GI-005-M)

[TOP](#)

---

<a name="GI-022-M"></a>
> 序号:GI-022-M | 接口描述：Topic发送群通知 | 传输协议: TCP

<small>
发送数据标识符：send_notify_topic
回执数据标识符：resp_notify_topic
</small>
该命令为topic发送消息的别名
用法参见：[GI-012-M Topic发送群消息](#GI-012-M)

[TOP](#)

---

<a name="GI-023-M"></a>
> 序号:GI-023-M | 接口描述：Topic接收群通知 | 传输协议: TCP

<small>
发送数据标识符：receive_notify_topic
回执数据标识符：topic_ack
</small>
该命令为topic接收消息的别名
用法参见：[GI-013-M Topic接收群消息](#GI-013-M)

[TOP](#)

---