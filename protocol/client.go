package protocol

//---------------------------------------------------------------------------
// 遵循客户端消息类型
//---------------------------------------------------------------------------
const (
	//好友上线下线 {notifyCode:6000,notifyMsg:"bb",type:"NOTIFY"}
	CLIENT_NOTIFY_FRIEND_ONLNE   = 6000
	CLIENT_NOTIFY_FRIEND_OFFLINE = 6001
)

type ClientNotifyMsg struct {
	NotifyCode int    `json:"notifyCode"`
	NotifyMsg  string `json:"notifyMsg"`
	Type       string `json:"type"`
}

func NewClientNotifyMsg(notifyCode int, notifyMsg string) *ClientNotifyMsg {
	return &ClientNotifyMsg{
		NotifyCode: notifyCode,
		NotifyMsg:  notifyMsg,
		Type:       "NOTIFY",
	}
}
