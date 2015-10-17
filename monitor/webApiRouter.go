package main

import (
	"encoding/json"
	"fmt"
	// "html"
	"goProject/log"
	"net/http"
)

//MsgServer
func ApiMsgServer(res http.ResponseWriter, req *http.Request) {

	msStr, err := json.Marshal(NewestMsgServerInfoData)
	if err != nil {
		log.Error(err.Error())
		return
	}

	fmt.Fprintf(res, string(msStr))
}
