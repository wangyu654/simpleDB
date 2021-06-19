package kvdb

func (ck *Clerk) Get(key string) string {
	// 整个客户端串行
	ck.mu.Lock()
	defer ck.mu.Unlock()

	args := &GetArgs{
		Key:       key,
		ClientId:  ck.clientId,
		RequestId: ck.requestId,
	}
	ck.requestId++
	for i := ck.leader; true; i = (i + 1) % len(ck.servers) {
		reply := &GetReply{}
		server := ck.servers[i]
		err := server.Call("KVServer.Get", args, reply)
		if err == nil {
			switch reply.Err {
			case ErrWrongLeader:
				continue
			case ErrNoKey:
				return ""
			case OK:
				ck.leader = i
				return reply.Value
			}
		}
	}
	return ""
}

func (ck *Clerk) PutAppend(key string, value string, op string) {
	ck.mu.Lock()
	defer ck.mu.Unlock()

	args := &PutAppendArgs{
		Key:       key,
		Value:     value,
		Op:        op,
		ClientId:  ck.clientId,
		RequestId: ck.requestId,
	}
	ck.requestId++
	for i := ck.leader; true; i = (i + 1) % len(ck.servers) {
		reply := &GetReply{}
		server := ck.servers[i]
		err := server.Call("KVServer.PutAppend", args, reply)
		if err == nil {
			switch reply.Err {
			case ErrWrongLeader:
				continue
			case OK:
				ck.leader = i
				return
			}
		}
	}

}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
