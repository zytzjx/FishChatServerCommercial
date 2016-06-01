package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	// "encoding/pem"
	"errors"
	"goProject/log"
	"io/ioutil"
	"net/http"
)

type Token struct {
	Data  TokenData
	Host  string
	Token string
}

type TokenData struct {
	I string //令牌ID
	A int64  //时间
	P string //存储路径
	R string //压缩参数
	T string //类型：image|vox
	C string //操作类型：add|del
}

func NewToken(data TokenData) *Token {
	var token Token
	token.Data = data
	token.GetFileServer()
	token.RsaEncrypt()
	return &token
}

// 加密
func (self *Token) RsaEncrypt() error {
	var (
		err  error
		temp []byte
	)

	temp, err = base64.StdEncoding.DecodeString(TOKEN_PUBLIC_KEY)
	if err != nil {
		log.Error("Error:", err)
		return err
	}

	pub, err := x509.ParsePKIXPublicKey(temp)
	if err != nil {
		log.Error("Error:", err)
		return err
	}

	temp, err = json.Marshal(self.Data)
	if err != nil {
		log.Error("Error:", err)
		return err
	}

	out, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), temp)
	if err != nil {
		log.Error("Error:", err)
		return err
	}

	self.Token = base64.StdEncoding.EncodeToString(out)
	return nil
}

// // 解密
// func (self *Token) RsaDecrypt() error {

// }

func (self *Token) GetFileServer() error {
	var (
		err     error
		resp    *http.Response
		temp    []byte
		servers []string
	)
	resp, err = http.Get(TOKEN_DISPATCH_URL)
	if err != nil {
		log.Error("Error:", err)
		return err
	}

	defer resp.Body.Close()
	temp, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error:", err)
		return err
	}

	err = json.Unmarshal(temp, &servers)
	if err != nil {
		log.Error("error:", err)
		return err
	}

	if len(servers) == 0 {
		err = errors.New("No file server can be find.")
		log.Error(err)
		return err
	}

	self.Host = servers[0]
	return nil
}
