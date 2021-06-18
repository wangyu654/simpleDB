package kvdb

import (
	"time"
)

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	cmd := Command{
		OpType: "Get",
		Args:   *args,
	}
	//提交到raft层 leader
	reply.Err = ErrWrongLeader
	index, _, isLeader := kv.rf.Start(cmd)
	if !isLeader {
		// 不是leader
		return
	}

	kv.mu.Lock()
	if _, ok := kv.messages[index]; !ok {
		kv.messages[index] = make(chan Message, 1)
	}
	chanMsg := kv.messages[index]
	kv.mu.Unlock()

	select {
	case msg := <-chanMsg:
		if recArgs, ok := msg.args.(GetArgs); !ok {
			return
		} else {
			if args.RequestId == recArgs.RequestId && args.ClientId == recArgs.ClientId {
				//校验 防止reply错误客户端
				*reply = msg.reply.(GetReply)
				// fmt.Printf("get %+v\n", reply)
				// fmt.Printf("get%+v\n", kv.database)
			}
		}
	case <-time.After(time.Second * 1):
	}
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	cmd := Command{
		OpType: "PutAppend",
		Args:   *args,
	}
	//提交到raft层 leader
	reply.Err = ErrWrongLeader
	index, _, isLeader := kv.rf.Start(cmd)
	if !isLeader {
		// 不是leader
		return
	}

	kv.mu.Lock()
	if _, ok := kv.messages[index]; !ok {
		kv.messages[index] = make(chan Message, 1)
	}
	chanMsg := kv.messages[index]
	kv.mu.Unlock()

	select {
	case msg := <-chanMsg:
		if recArgs, ok := msg.args.(PutAppendArgs); !ok {
			return
		} else {
			if args.RequestId == recArgs.RequestId && args.ClientId == recArgs.ClientId {
				//校验 防止reply错误客户端
				*reply = msg.reply.(PutAppendReply)
				// fmt.Printf("put %+v\n", reply)
				// fmt.Printf("put %+v\n", kv.database)
			}
		}
	case <-time.After(time.Second * 1):
	}
}
