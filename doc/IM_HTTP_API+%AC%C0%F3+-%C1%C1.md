
# GOLANG IM HTTP API 开发文档

> <small>nprog | <ingram@60.com> | 2016/4/21 | version: 0.1.1</small>


###目录

1. [GH-001 请求P2P历史记录接口](#GH-001)
2. [GH-002 请求Topic历史记录接口](#GH-002)
3. [GH-003 注册用户](#GH-003)
4. [GH-004 查看在线好友列表](#GH-004)
5. [GH-005 添加好友](#GH-005)
6. [GH-006 获取离线信息](#GH-006)


###更新日志
> 
>  * 2015/11/2 | version: 0.1
>  <small>新增查询P2P历史纪录, Topic历史纪录</small>
>  * 2016/4/21 | version: 0.1.1
>  <small>新增未读信息拉取</small>


###文档

> 线下测试服务
http://192.168.60.119:28080
> 线上测试服务
http://gw.60talk.com:28080

<a name="GH-001"></a>
> 序号:GH-001 | 接口描述：请求P2P历史记录接口 | 请求类型: HTTP/GET

URL: /history/v1/p2pHistory

><small>传入参数</small>

| Filed | type | require | default | description |
|:---- |:---- |:---- |:---- |:---- |
| token | String | true | | token验证方式，现在直接传入用户ID |
| friendId | String | true | | 要查询的好友ID |
| endTime | String | | Now | 从endTime往前查msgNum条数据 |
| msgNum | Integer | | 100 | 分页最大返回条数 |

<small>
> 返回数据示例:

```
/history/v1/p2pHistory?token=aa&friendId=b
{
    "status": "0000",
    "result": {
        "total": 2,
        "pageSize": 2,
        "endTime": 1446451206,
        "data": [
            {
                "msgType": "send_message_p2p",
                "fromId": "aa",
                "friendId": "bb",
                "content": "{\"fromUser\":\"aa\",\"message\":\"看咯莫\",\"toUser\":\"bb\",\"type\":\"TXT\"}",
                "time": 1445910961,
                "uuid": "d7ae90cd-946f-4837-ad9c-b17d5c4e3085"
            },
            {
                "msgType": "send_message_p2p",
                "fromId": "aa",
                "friendId": "bb",
                "content": "{\"fromUser\":\"aa\",\"message\":\"可口可乐了\",\"toUser\":\"bb\",\"type\":\"TXT\"}",
                "time": 1445860312,
                "uuid": "41844271-7de6-4fa2-b6ca-73e47262433c"
            }
        ]
    }
}
```
</small>


[TOP](#)

---

<a name="GH-002"></a>
> 序号:GH-002 | 接口描述：请求Topic历史记录接口 | 请求类型: HTTP/GET

URL:/history/v1/topicHistory

><small>传入参数</small>

| Filed | type | require | default | description |
|:---- |:---- |:---- |:---- |:---- |
| token | String | true | | token验证方式，现在直接传入用户ID |
| topicId | String | true | | 要查询的群组ID |
| endTime | String | | Now | 从endTime往前查msgNum条数据 |
| msgNum | Integer | | 100 | 分页最大返回条数 |

<small>
> 返回数据示例:

```
/history/v1/topicHistory?token=qingshanz&topicId=60talk_topic&msgNum=2
{
    "status": "0000",
    "result": {
        "total": 2,
        "pageSize": 2,
        "endTime": 1446452799,
        "data": [
            {
                "msgType": "send_message_topic",
                "fromId": "qingshanz",
                "topicId": "60talk_topic",
                "content": "{\"fromUser\":\"qingshanz\",\"groupName\":\"60talk_topic\",\"message\":\"你好 60talk\",\"type\":\"TXT\"}",
                "time": 1445668876,
                "uuid": "97834f33-3059-4ce9-872b-a3067cbb153a"
            },
            {
                "msgType": "send_message_topic",
                "fromId": "qingshanz",
                "topicId": "60talk_topic",
                "content": "{\"fromUser\":\"qingshanz\",\"groupName\":\"60talk_topic\",\"message\":\"你好 60talk\",\"type\":\"TXT\"}",
                "time": 1445668785,
                "uuid": "3a3d6326-be9a-494a-b33f-f3232aaa8c6f"
            }
        ]
    }
}
```
</small>

[TOP](#)

---

<a name="GH-003"></a>
> 序号:GH-003 | 接口描述：注册用户 | 请求类型: HTTP/GET

URL:/user/v1/register

><small>传入参数</small>

| Filed | type | require | default | description |
|:---- |:---- |:---- |:---- |:---- |
| clientId | String | true | | 用户ID |

<small>
> 返回数据示例:

```
/user/v1/register?clientId=qingshanz
{
    "status": "0000",
    "result": {}
}
```
</small>

[TOP](#)

---

<a name="GH-004"></a>
> 序号:GH-004 | 接口描述：查看在线好友列表 | 请求类型: HTTP/GET

URL:/friend/v1/viewFriend

><small>传入参数</small>

| Filed | type | require | default | description |
|:---- |:---- |:---- |:---- |:---- |
| cid | String | true | | 用户ID |

<small>
> 返回数据示例:

```
/user/v1/register?cid=qingshanz
{
    "status": "0000",
    "result": {
        "friends_alive":["aa"]
    }
}
```
</small>

[TOP](#)

---


<a name="GH-005"></a>
> 序号:GH-005 | 接口描述：添加好友 | 请求类型: HTTP/GET

URL:/friend/v1/addFriend

><small>传入参数</small>

| Filed | type | require | default | description |
|:---- |:---- |:---- |:---- |:---- |
| cid | String | true | | 用户ID |
| friendId | String | true | | 用户ID |

<small>
> 返回数据示例:

```
/friend/v1/addFriend?cid=bb&friendId=aa
{
    "status": "0000",
    "result": {}
}
```
</small>

[TOP](#)

---

<a name="GH-006"></a>
> 序号:GH-006 | 接口描述：获取离线信息 | 请求类型: HTTP/GET

URL:/pull/v1/p2pMsg

><small>传入参数</small>

| Filed | type | require | default | description |
|:---- |:---- |:---- |:---- |:---- |
| token | String | true | | 用户ID |

<small>
> 返回数据示例:

```
/user/v1/register?token=286
{
    "status": "0000",
    "result": {
        "total":1,
        "data":[{
            "msgType":"send_message_p2p",
            "fromId":"1034",
            "toId":"286",
            "content":"{\n  \"fromUser\" : \"1034\",\n  \"toUser\" : \"286\",\n  \"message\" : \"解决了\",\n  \"type\" : \"TXT\"\n}",
            "time":1446100432,
            "uuid":"cf3126c4-7a1c-4c9b-a6f5-2bad4ef8c9f7"
        }]
    }
}
```
</small>

> <small>NOTE: 本接口已经打包的数据已经标记成已读，不用再回ACK</small>

[TOP](#)

---
