package service_discovery

import (
	"encoding/json"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"goProject/common"
	"goProject/log"
	"time"
)

type Master struct {
	Members map[string]*Member
	KeysAPI client.KeysAPI
}

// Member is a client machine
type Member struct {
	InGroup    bool
	IP         string
	Name       string
	CPU        int
	SessionNum uint64
}

func NewMaster(endpoints []string) *Master {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		log.Fatal("Error: cannot connec to etcd:", err)
	}

	master := &Master{
		Members: make(map[string]*Member),
		KeysAPI: client.NewKeysAPI(etcdClient),
	}
	go master.WatchWorkers()
	return master
}

func (m *Master) AddWorker(info *WorkerInfo) {
	member := &Member{
		InGroup:    true,
		IP:         info.IP,
		Name:       info.Name,
		CPU:        info.CPU,
		SessionNum: info.SessionNum,
	}
	m.Members[member.Name] = member
}

func (m *Master) UpdateWorker(info *WorkerInfo) {
	member := &Member{
		InGroup:    true,
		IP:         info.IP,
		Name:       info.Name,
		CPU:        info.CPU,
		SessionNum: info.SessionNum,
	}
	m.Members[info.Name] = member
}

func (m *Master) WatchWorkers() {
	api := m.KeysAPI
	watcher := api.Watcher("workers/", &client.WatcherOptions{
		Recursive: true,
	})
	for {
		res, err := watcher.Next(context.Background())
		if err != nil {
			log.Error("Error watch workers:", err.Error())
			break
		}

		if res.Action == "expire" {
			key := common.Substr(res.Node.Key, len("/workers/"), len(res.Node.Key)-len("/workers/"))
			member, ok := m.Members[key]
			if ok {
				member.InGroup = false
			}
		} else if res.Action == "set" || res.Action == "update" {
			info := WorkerInfo{}
			err := json.Unmarshal([]byte(res.Node.Value), &info)
			if err != nil {
				log.Error(err.Error())
			}
			// log.Info(info)
			if _, ok := m.Members[info.Name]; ok {
				m.UpdateWorker(&info)
			} else {
				m.AddWorker(&info)
			}
		} else if res.Action == "delete" {
			delete(m.Members, res.Node.Key)
		}
	}

}
