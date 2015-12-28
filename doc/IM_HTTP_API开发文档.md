
# GOLANG IM HTTP API 开发文档

> <small>nprog | <ingram@60.com> | 2015/11/2 | version: 0.1</small>


###目录

1. [GH-001 请求P2P历史记录接口](#GH-001)
2. [GH-002 请求Topic历史记录接口](#GH-002)


###更新日志
> 
>  * 2015/11/2 | version: 0.1
>  <small>新增查询P2P历史纪录, Topic历史纪录</small>


###文档

<a name="GH-001"></a>
> 序号:GH-001 | 接口描述：请求P2P历史记录接口 | 请求类型: HTTP/GET

URL: http://192.168.60.119:28080/history/v1/p2pHistory

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
http://192.168.60.119:28080/history/v1/p2pHistory?token=aa&friendId=b
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

URL:http://192.168.60.119:28080/history/v1/topicHistory

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
http://192.168.60.119:28080/history/v1/topicHistory?token=qingshanz&topicId=60talk_topic&msgNum=2
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

