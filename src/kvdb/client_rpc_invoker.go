package kvdb

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
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
		if server.Call("KVServer.Get", args, reply) {
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

//
// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
//
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
		if server.Call("KVServer.PutAppend", args, reply) {
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
