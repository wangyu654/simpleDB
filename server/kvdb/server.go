package kvdb

import (
	"simpleDB/server/engine"
	"simpleDB/server/raft"
	"strconv"
	"sync"
)

type Command struct {
	OpType string
	Args   interface{} //PutAppendArgs or GetArgs
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg
	dead    int32

	maxraftstate int // snapshot if log grows this big

	messages map[int]chan Message
	ack      map[int64]int
	database engine.Engine
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
		kv.printInfo("get log from raft:", raftLog)
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

// 发送消息
func (kv *KVServer) sendMsg(index int, msg Message) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if _, ok := kv.messages[index]; !ok {
		kv.messages[index] = make(chan Message, 1)
	}

	kv.messages[index] <- msg
}

// 应用日志
func (kv *KVServer) apply(command Command, isDuplicated bool) interface{} {
	var (
		key uint64
		err error
	)

	kv.mu.Lock()
	defer kv.mu.Unlock()
	switch command.OpType {
	case "Get":
		args := command.Args.(GetArgs)
		reply := GetReply{}
		if key, err = strconv.ParseUint(args.Key, 10, 64); err != nil {
			reply.Err = ErrNoKey
			return reply
		}
		if value, err := kv.database.Find(key); err == nil {
			reply.Err = OK
			reply.Value = value
		} else {
			reply.Err = ErrNoKey
		}
		return reply
	case "PutAppend":
		args := command.Args.(PutAppendArgs)
		reply := PutAppendReply{}
		if key, err = strconv.ParseUint(args.Key, 10, 64); err != nil {
			reply.Err = ErrNoKey
			return reply
		}
		if !isDuplicated {
			if args.Op == "Put" {
				kv.database.Insert(key, args.Value)
			} else if args.Op == "Append" {
				if value, err := kv.database.Find(key); err == nil {
					kv.database.Update(key, value+args.Value)
				} else {
					reply.Err = ErrNoKey
				}
			}
		}
		reply.Err = OK
		return reply
	}
	return nil
}

// 判定重复
func (kv *KVServer) isDuplicated(clientId int64, requestId int) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if maxRequestId, ok := kv.ack[clientId]; ok && requestId <= maxRequestId {
		return true
	}
	kv.ack[clientId] = requestId
	return false
}
