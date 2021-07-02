package raft

import (
	"net/rpc"
	"time"

	"github.com/ngaut/log"
)

const retryTime = 1 * time.Second

func (rf *Raft) Sysmon() {

	rf.mu.Lock()
	defer rf.mu.Unlock()

	for i := range rf.peers {
		server := rf.peers[i]
		if !rf.loseConn[i] && server.Close().Error() != "" {
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
