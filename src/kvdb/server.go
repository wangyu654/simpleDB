package kvdb

import (
	"golab1/src/labgob"
	"golab1/src/labrpc"
	"golab1/src/raft"
	"log"
	"sync"
	"sync/atomic"
)

const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

type Command struct {
	OpType string
	Args   interface{}
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate int // snapshot if log grows this big

	messages map[int]chan Message
	ack      map[int64]int
	database map[string]string // for storing data

}

func (kv *KVServer) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
	// Your code here, if desired.
}

func (kv *KVServer) killed() bool {
	z := atomic.LoadInt32(&kv.dead)
	return z == 1
}

func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	labgob.Register(Command{})
	labgob.Register(PutAppendArgs{})
	labgob.Register(GetArgs{})
	labgob.Register(PutAppendReply{})
	labgob.Register(GetReply{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.ack = map[int64]int{}
	kv.database = map[string]string{}
	kv.messages = map[int]chan Message{}

	go kv.getLogFromRaft()

	return kv
}

type Message struct {
	opType string
	args   interface{}
	reply  interface{}
}

func (kv *KVServer) getLogFromRaft() {
	for {
		// 在raft层拿到log
		raftLog := <-kv.applyCh
		command := raftLog.Command.(Command) //类型转换
		var msg Message
		var clientId int64
		var requestId int
		if command.OpType == "Get" {
			// get请求
			args := command.Args.(GetArgs)
			clientId = args.ClientId
			requestId = args.RequestId
			msg.args = args
		} else {
			// putAppend请求
			args := command.Args.(PutAppendArgs)
			clientId = args.ClientId
			requestId = args.RequestId
			msg.args = args
		}
		msg.opType = command.OpType
		msg.reply = kv.apply(command, kv.isDuplicated(clientId, requestId))
		kv.sendMsg(raftLog.CommandIndex, msg)
	}
}

func (kv *KVServer) sendMsg(index int, msg Message) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if _, ok := kv.messages[index]; !ok {
		kv.messages[index] = make(chan Message, 1)
	}

	kv.messages[index] <- msg
}

func (kv *KVServer) apply(command Command, isDuplicated bool) interface{} {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	// fmt.Println("command:", command)
	switch command.OpType {
	case "Get":
		args := command.Args.(GetArgs)
		kv.printInfo("get", args.Key)
		reply := GetReply{}
		if value, ok := kv.database[args.Key]; ok {
			reply.Err = OK
			reply.Value = value
		} else {
			reply.Err = ErrNoKey
		}
		return reply
	case "PutAppend":
		args := command.Args.(PutAppendArgs)
		// fmt.Println("args:", args, "reqeustId", args.RequestId)
		reply := PutAppendReply{}
		if !isDuplicated {
			if args.Op == "Put" {
				// kv.printInfo("put key:", args.Key, "value:", args.Value)
				kv.database[args.Key] = args.Value
			} else if args.Op == "Append" {
				kv.database[args.Key] += args.Value
				// kv.printInfo("append key:", args.Key, "value:", kv.database[args.Key])
			}
		}
		reply.Err = OK
		return reply
	}
	return nil
}

func (kv *KVServer) isDuplicated(clientId int64, requestId int) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if maxRequestId, ok := kv.ack[clientId]; ok && requestId <= maxRequestId {
		return true
	}
	kv.ack[clientId] = requestId
	return false
}
