package client

import (
	"crypto/rand"
	"math/big"
	"net/rpc"
	"sync"
)

type Clerk struct {
	servers   []*rpc.Client
	mu        sync.Mutex
	leader    int
	requestId int //递增
	clientId  int64
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*rpc.Client) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.clientId = nrand()
	ck.mu = sync.Mutex{}
	return ck
}
