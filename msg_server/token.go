package main

import (
	"goProject/base"
	"goProject/common"
	"goProject/info"
	"goProject/libnet"
	"goProject/log"
	"goProject/protocol"
	"goProject/token"
	"strconv"
	"strings"
	"time"
)

//router订阅请求
func (self *ProtoProc) procSendGetToken(cmd protocol.Cmd, session *libnet.Session) error {
	log.Info("procSendGetToken")
	var (
		err                error
		fileName           string
		exTime             int64
		rootPath           string
		path               string
		compressOption     string
		resType            string
		actionType         string
		clientIdAndAppName string
		clientId           string
		appName            string
		tk                 *token.Token
	)
	if session.State == nil {
		self.respCmd(protocol.RESP_GET_TOKEN, session, cmd.GetReport(), false, info.YOU_HAVE_NOT_LANDED)
		return nil
	}

	if len(cmd.GetArgs()) < protocol.SEND_GET_TOKEN_ARGS_NUM {
		log.Info(info.NOT_ENOUGH_ARGUMENTS)
		self.respCmd(protocol.RESP_GET_TOKEN, session, cmd.GetReport(), false, info.NOT_ENOUGH_ARGUMENTS)
		return nil
	}

	if len(cmd.GetArgs()) > 1 {
		actionType = cmd.GetArgs()[1]
	} else {
		actionType = token.TOKEN_ACTION_TYPE_ADD
	}

	if len(cmd.GetArgs()) > 2 {
		compressOption = cmd.GetArgs()[2]
	}

	clientIdAndAppName = session.State.(*base.SessionState).ClientID

	clientIdAndAppNameArr := strings.Split(clientIdAndAppName, "#")
	if len(clientIdAndAppNameArr) > 0 {
		clientId = clientIdAndAppNameArr[0]
	} else {
		clientId = "defaultClient"
	}
	if len(clientIdAndAppNameArr) > 1 {
		appName = clientIdAndAppNameArr[1]
	} else {
		appName = "defaultApp"
	}

	resType = cmd.GetArgs()[0]
	exTime = time.Now().Unix()
	fileName = common.NewV4().String()[0:8]

	path = appName + "/" +
		clientId + "/" +
		strconv.Itoa(time.Now().Year()) +
		strconv.Itoa(int(time.Now().Month())) +
		strconv.Itoa(time.Now().Day()) + "/"

	tk = token.NewToken(token.TokenData{
		I: fileName,
		A: exTime,
		P: path,
		R: compressOption,
		T: resType,
		C: actionType,
	})

	var (
		doMain string
		upUrl  string
	)

	upUrl = "http://" + tk.Host + ":10060/v1/"
	//先写死后缀
	switch resType {
	case token.TOKEN_TYPE_IMG:
		fileName = fileName + ".jpg"
		rootPath = "/images/"
		upUrl = upUrl + "image?method=add"
	case token.TOKEN_TYPE_VOX:
		fileName = fileName + ".amr"
		rootPath = "/vox/"
		upUrl = upUrl + "vox?method=add"
	case token.TOKEN_TYPE_FILE:
		fileName = ""
		rootPath = ""
		upUrl = upUrl + "file?method=add"
	}

	doMain = "http://" + tk.Host + rootPath + path + fileName

	// //RESP_GET_TOKEN TOKEN FILENAME PATH DOMAIN UPURL
	resp := protocol.NewCmdResponse(protocol.RESP_GET_TOKEN)
	resp.Time = time.Now().Unix()
	resp.Repo = cmd.GetReport()
	resp.AddArg(tk.Token)
	resp.AddArg(fileName)
	resp.AddArg(rootPath + path)
	resp.AddArg(doMain)
	resp.AddArg(upUrl)

	//返回用户请求
	err = session.Send(resp)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}
