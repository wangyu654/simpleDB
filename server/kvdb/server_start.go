package kvdb

import (
	"encoding/gob"
	"log"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"simpleDB/server/engine/bptree"
	"simpleDB/server/raft"
)

type Client struct {
	index int
	conn  *rpc.Client
}

func StartKVServer(servers []*rpc.Client, me int, persister *raft.Persister, maxraftstate int, address string, ch chan bool) {

	var err error
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	kv := new(KVServer)
	rpc.Register(kv)
	gob.Register(Command{})
	gob.Register(GetArgs{})
	gob.Register(GetReply{})
	gob.Register(PutAppendArgs{})
	gob.Register(PutAppendReply{})

	kv.me = me
	kv.maxraftstate = maxraftstate

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.ack = map[int64]int{}
	engine, err := bptree.NewTree("db" + strconv.Itoa(me) + ".wy")
	if err != nil {
		panic(err)
	}
	kv.database = engine
	kv.messages = map[int]chan Message{}

	go func() {
		log.Println("waiting for raft prepared")
		<-ch
		kv.rf = raft.Make(servers, me, persister, kv.applyCh)
		log.Println("raft prepared")
		kv.getLogFromRaft()
	}()

	rpc.HandleHTTP()
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Panicln(err)
	}

}

func getConnection(addresses []string, index int, servers *[]*rpc.Client, finished chan bool) {
	ch := make(chan *Client)
	for i, address := range addresses {
		if i == index {
			continue
		}
		go func(address string, ch chan *Client, i int) {
			conn, err := rpc.DialHTTP("tcp", address)
			for err != nil {
				log.Println("try to connect", address)
				time.Sleep(time.Duration(time.Second))
				conn, err = rpc.DialHTTP("tcp", address)
			}
			ch <- &Client{
				index: i,
				conn:  conn,
			}
		}(address, ch, i)
	}
	rest := len(addresses)
	for i := 0; i < rest; i++ {
		c := <-ch
		(*servers)[c.index] = c.conn
		rest--
	}
	finished <- true
}

func Start(addresses []string, index int) {
	servers := make([]*rpc.Client, len(addresses))
	ch := make(chan bool)
	go getConnection(addresses, index, &servers, ch)
	StartKVServer(servers, index, raft.MakePersister(), -1, addresses[index], ch)
}
