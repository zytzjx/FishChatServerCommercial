package main

import (
	"encoding/json"
	"errors"
	"goProject/common"
	"goProject/log"
	"goProject/storage/mongo_store"
	"net/http"
	"strconv"
	"time"
)

type handle struct {
	Db *mongo_store.MongoStore
}

func NewHandle(db *mongo_store.MongoStore) *handle {
	return &handle{
		Db: db,
	}
}

func (self *handle) Route(w http.ResponseWriter, r *http.Request) {
	log.Info(r.URL.Path)

	w.Header().Set("content-type", "application/json")
	switch r.URL.Path {
	case ROUTER_HOME:
		self.Home(w, r)
	case ROUTER_P2P_HISTORY:
		self.P2PHistory(w, r)
	case ROUTER_TOPIC_HISTORY:
		self.TopicHistory(w, r)
	case ROUTER_VIEW_FRIEND:
		self.ViewFriend(w, r)
	case ROUTER_ADD_FRIEND:
		self.AddFriend(w, r)
	case ROUTER_USER_REGISTER:
		self.Register(w, r)
	default:
		w.Write([]byte("404 page not find"))
	}
}

func (self *handle) Home(w http.ResponseWriter, r *http.Request) {

}

func (self *handle) P2PHistory(w http.ResponseWriter, r *http.Request) {
	log.Info("::P2PHistory")
	var (
		err     error
		fromID  string
		toID    string
		endTime int64
		n       int
		resp    BaseResultTemple
		emp     EmptyTemple
	)

	if self.GetParam(r, "token") != "" {
		fromID = self.GetParam(r, "token")
	}
	if self.GetParam(r, "friendId") != "" {
		toID = self.GetParam(r, "friendId")
	}
	if self.GetParam(r, "endTime") != "" {
		endTime, err = strconv.ParseInt(self.GetParam(r, "endTime"), 10, 64)
		if err != nil {
			log.Error(err.Error())
			resp.Status = RESP_STATUS_ERROR
			resp.Result = emp
			self.Response(w, resp)
			return
		}
	} else {
		endTime = time.Now().Unix()
	}
	if self.GetParam(r, "msgNum") != "" {
		n, err = strconv.Atoi(self.GetParam(r, "msgNum"))
		if err != nil {
			log.Info("msgNum: ", err)
			resp.Status = RESP_STATUS_ERROR
			resp.Result = emp
			self.Response(w, resp)
			return
		}
	} else {
		n = DEFAULT_GET_MSG_NUM
	}

	if fromID == "" || toID == "" {
		log.Info("need fromid or toid.")
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	var (
		mrt  MsgsResultTemple
		data []P2PMsgTemple
	)
	mrt.PageSize = n
	mrt.EndTime = endTime

	result := self.Db.ReadP2PHistoryFromEndTime(mongo_store.DATA_BASE_NAME,
		mongo_store.RECORD_P2P_MESSAGE_COLLECTION, fromID, toID, endTime, n)

	if len(result) > 0 {
		for i := 0; i < len(result); i++ {
			data = append(data, P2PMsgTemple{
				MsgType:  result[i].MsgType,
				FromID:   result[i].FromID,
				FriendId: result[i].ToID,
				Content:  result[i].Content,
				Time:     result[i].Time,
				UUID:     result[i].UUID,
			})
		}
		mrt.Data = data
		mrt.Total = len(result)
	} else {
		mrt.Data = emp
	}

	resp.Status = RESP_STATUS_SUCCESS
	resp.Result = mrt
	err = self.Response(w, resp)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func (self *handle) TopicHistory(w http.ResponseWriter, r *http.Request) {
	log.Info("::TopicHistory")

	var (
		err     error
		fromID  string
		TopicID string
		endTime int64
		n       int
		resp    BaseResultTemple
		emp     EmptyTemple
	)

	if self.GetParam(r, "token") != "" {
		fromID = self.GetParam(r, "token")
	}
	if self.GetParam(r, "topicId") != "" {
		TopicID = self.GetParam(r, "topicId")
	}
	if self.GetParam(r, "endTime") != "" {
		endTime, err = strconv.ParseInt(self.GetParam(r, "endTime"), 10, 64)
		if err != nil {
			log.Error(err.Error())
			resp.Status = RESP_STATUS_ERROR
			resp.Result = emp
			self.Response(w, resp)
			return
		}
	} else {
		endTime = time.Now().Unix()
	}
	if self.GetParam(r, "msgNum") != "" {
		n, err = strconv.Atoi(self.GetParam(r, "msgNum"))
		if err != nil {
			log.Info("msgNum: ", err)
			resp.Status = RESP_STATUS_ERROR
			resp.Result = emp
			self.Response(w, resp)
			return
		}
	} else {
		n = DEFAULT_GET_MSG_NUM
	}

	if fromID == "" || TopicID == "" {
		log.Info("need fromid or toid.")
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	var (
		mrt  MsgsResultTemple
		data []TopicMsgTemple
	)
	mrt.PageSize = n
	mrt.EndTime = endTime

	result := self.Db.ReadTopicHistoryFromEndTime(mongo_store.DATA_BASE_NAME,
		mongo_store.RECORD_TOPIC_MESSAGE_COLLECTION, TopicID, endTime, n)

	if len(result) > 0 {
		for i := 0; i < len(result); i++ {
			data = append(data, TopicMsgTemple{
				MsgType: result[i].MsgType,
				FromID:  result[i].FromID,
				TopicID: result[i].ToID,
				Content: result[i].Content,
				Time:    result[i].Time,
				UUID:    result[i].UUID,
			})
		}
		mrt.Data = data
		mrt.Total = len(result)
	} else {
		mrt.Data = emp
	}

	resp.Status = RESP_STATUS_SUCCESS
	resp.Result = mrt
	err = self.Response(w, resp)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func (self *handle) ViewFriend(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		cid  string
		resp BaseResultTemple
		emp  EmptyTemple
		frt  FriendAliveResultTemple
	)
	if self.GetParam(r, "cid") != "" {
		cid = self.GetParam(r, "cid")
	}

	clientInfo, err := self.Db.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, cid)
	if err != nil {
		log.Error(err.Error())
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	friends := clientInfo.Friends
	if friends != nil {
		for _, value := range friends {
			clientInfo, err = self.Db.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, value)
			if err != nil {
				log.Error(err.Error())
				resp.Status = RESP_STATUS_ERROR
				resp.Result = emp
				self.Response(w, resp)
				return
			}
			// tmpSingle := make(map[string]bool)
			// tmpSingle[value] = clientInfo.Alive

			// frt.FriendAlive[value] = clientInfo.Alive
			if clientInfo.Alive {
				frt.FriendAlive = append(frt.FriendAlive, value)
			}

		}

	}

	resp.Status = RESP_STATUS_SUCCESS
	resp.Result = frt
	err = self.Response(w, resp)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func (self *handle) AddFriend(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		cid      string
		friendId string
		resp     BaseResultTemple
		emp      EmptyTemple
	)
	if self.GetParam(r, "cid") != "" {
		cid = self.GetParam(r, "cid")
	}

	if self.GetParam(r, "friendId") != "" {
		friendId = self.GetParam(r, "friendId")
	}

	clientInfo, err := self.Db.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, cid)
	if err != nil {
		log.Error(err.Error())
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	friendInfo, err := self.Db.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, friendId)
	if err != nil {
		notFound := errors.New("not found")
		if err.Error() == notFound.Error() {
			log.Error(err.Error())
			resp.Status = RESP_STATUS_ERROR
			resp.Result = emp
			self.Response(w, resp)
			return
		} else {
			log.Error(err.Error())
			resp.Status = RESP_STATUS_ERROR
			resp.Result = emp
			self.Response(w, resp)
			return
		}
	}

	friends := clientInfo.Friends
	if common.InArray(friends, friendId) {
		resp.Status = RESP_STATUS_SUCCESS
		resp.Result = FriendAliveTemple{friendInfo.ClientID, friendInfo.Alive}
		self.Response(w, resp)
		return
	}

	err = self.Db.UpdateFriendsFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, cid, append(friends, friendId))
	if err != nil {
		log.Error(err.Error())
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	resp.Status = RESP_STATUS_SUCCESS
	resp.Result = FriendAliveTemple{friendInfo.ClientID, friendInfo.Alive}
	err = self.Response(w, resp)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func (self *handle) Register(w http.ResponseWriter, r *http.Request) {
	log.Info("::Register")
	var (
		err      error
		clientId string
		resp     BaseResultTemple
		emp      EmptyTemple
	)

	if self.GetParam(r, "clientId") != "" {
		clientId = self.GetParam(r, "clientId")
	}

	if clientId == "" {
		log.Error("need clientId.")
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	//查找用户信息
	clientInfo, err := self.Db.GetClientFromId(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, clientId)
	if clientInfo != nil {
		resp.Status = RESP_STATUS_REPEAT_REGISTRATION
		resp.Result = emp
		self.Response(w, resp)
		return
	}

	sessionStoreData := mongo_store.SessionStoreData{clientId, "",
		"", []string{}, false, ""}

	// update login info
	err = self.Db.Upsert(mongo_store.DATA_BASE_NAME, mongo_store.CLIENT_INFO_COLLECTION, &sessionStoreData)
	if err != nil {
		log.Error(err.Error())
		resp.Status = RESP_STATUS_ERROR
		resp.Result = emp
		self.Response(w, resp)
		return
	}
	resp.Status = RESP_STATUS_SUCCESS
	resp.Result = emp
	err = self.Response(w, resp)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func (self *handle) GetParam(r *http.Request, param string) string {
	r.ParseForm()
	if len(r.Form[param]) > 0 {
		return r.Form[param][0]
	}
	return ""
}

func (self *handle) Response(w http.ResponseWriter, resp BaseResultTemple) error {
	temp, err := json.Marshal(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	w.Write(temp)
	return nil
}
