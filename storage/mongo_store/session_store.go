
package mongo_store


type SessionStoreData struct {
	ClientID       string  `bson:"ClientID"`
	ClientAddr     string  `bson:"ClientAddr"`
	MsgServerAddr  string  `bson:"MsgServerAddr"`
	Alive          bool    `bson:"Alive"`
}
