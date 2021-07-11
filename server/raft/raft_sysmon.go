package raft

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/ngaut/log"
)

const retryTime = 1 * time.Second

func (rf *Raft) Sysmon() {

	rf.mu.Lock()
	defer rf.mu.Unlock()
	fmt.Println(rf.peers)
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		server := rf.peers[i]
		// fmt.Println(server.Close().Error())
		if !rf.loseConn[i] && server == nil {
			rf.loseConn[i] = true
			go rf.retry(i)
		}
	}

}

func (rf *Raft) retry(i int) {
	for {
		address := rf.addresses[i]
		conn, err := rpc.Dial("tcp", address)
		if err != nil {
			log.Error(err)
		} else {
			rf.mu.Lock()
			defer rf.mu.Unlock()
			rf.loseConn[i] = false
			rf.peers[i] = conn
			return
		}
		time.After(retryTime)
	}
}
