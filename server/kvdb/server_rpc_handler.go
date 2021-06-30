package kvdb

import (
	"time"
)

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) error {
	cmd := Command{
		OpType: "Get",
		Args:   *args,
	}
	//提交到raft层 leader
	reply.Err = OK
	index, _, isLeader := kv.rf.Start(cmd)
	if !isLeader {
		reply.Err = ErrWrongLeader
		// 不是leader
		return nil
	}
	kv.mu.Lock()
	if _, ok := kv.messages[index]; !ok {
		kv.messages[index] = make(chan Message, 1)
	}
	chanMsg := kv.messages[index]
	kv.mu.Unlock()

	select {
	case msg := <-chanMsg:
		// kv.printInfo("msg from :", msg, "!!!!")
		if recArgs, ok := msg.args.(GetArgs); !ok {
			return nil
		} else {
			if args.RequestId == recArgs.RequestId && args.ClientId == recArgs.ClientId {
				//校验 防止reply错误客户端
				*reply = msg.reply.(GetReply)
			}
		}
	case <-time.After(time.Second * 10):
	}
	return nil
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) error {
	cmd := Command{
		OpType: "PutAppend",
		Args:   *args,
	}
	//提交到raft层 leader
	reply.Err = OK

	index, _, isLeader := kv.rf.Start(cmd)
	if !isLeader {
		reply.Err = ErrWrongLeader
		// 不是leader
		return nil
	}
	kv.mu.Lock()
	if _, ok := kv.messages[index]; !ok {
		kv.messages[index] = make(chan Message, 1)
	}
	chanMsg := kv.messages[index]
	kv.mu.Unlock()

	select {
	case msg := <-chanMsg:
		kv.printInfo("receive cmd:", msg)
		if recArgs, ok := msg.args.(PutAppendArgs); !ok {
			return nil
		} else {
			if args.RequestId == recArgs.RequestId && args.ClientId == recArgs.ClientId {
				//校验 防止reply错误客户端
				*reply = msg.reply.(PutAppendReply)
			}
		}
	case <-time.After(time.Second * 1):
	}
	return nil
}
