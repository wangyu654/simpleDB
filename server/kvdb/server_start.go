package kvdb

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"

	"simpleDB/server/engine/bptree"
	"simpleDB/server/raft"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Index     int      `yaml:"index"`
	Addresses []string `yaml:"addresses"`
}

func readConfig() (*Config, error) {
	var err error

	config, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err = yaml.Unmarshal(config, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func StartKVServer(servers []*rpc.Client, me int, persister *raft.Persister, maxraftstate int, address string) {

	rpc.Register(new(raft.Raft))
	rpc.Register(new(KVServer))

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.ack = map[int64]int{}
	kv.database = &bptree.Tree{}
	kv.messages = map[int]chan Message{}

	go kv.getLogFromRaft()

	rpc.HandleHTTP()
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Panicln(err)
	}
}

func Start() {
	/*
		n := 0
		addresses := []string{}
		读取配置
	*/
	cfg, err := readConfig()
	if err != nil {
		panic("fail to read config")
	}
	servers := []*rpc.Client{}
	for i, address := range cfg.Addresses {
		if i == cfg.Index {
			continue
		}
		conn, err := rpc.DialHTTP("tcp", address)
		if err != nil {
			log.Fatal(err)
		}
		servers = append(servers, conn)
	}
	StartKVServer(servers, cfg.Index, raft.MakePersister(), -1, cfg.Addresses[cfg.Index])
}
