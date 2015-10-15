

package common

import (
	"math/rand"
	//"goProject/storage/mongo_store"
)


//Just use random to select msg_server
func SelectServer(serverList []string, serverNum int) string {
	return serverList[rand.Intn(serverNum)]
}




