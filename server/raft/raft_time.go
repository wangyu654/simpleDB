package raft

import (
	"math/rand"
	"time"
)

const RaftElectionTimeoutLow = 150 * time.Millisecond
const RaftElectionTimeoutHigh = 300 * time.Millisecond
const RaftHeartbeatPeriod = 50 * time.Millisecond

func (follower *Raft) startElectTimer() {
	floatInterval := int(RaftElectionTimeoutHigh - RaftElectionTimeoutLow)
	timeout := time.Duration(rand.Intn(floatInterval)) + RaftElectionTimeoutLow
	electTimer := time.NewTimer(timeout)
	for {
		select {
		case <-follower.chanGrantVote:
			// voted for other server
		case <-follower.chanHeartbeat:
			// receive heartbeat
		case <-electTimer.C:
			follower.printInfo("electTimer expired")
			follower.toCandidate() //选举超时 新term
		}
		floatInterval := int(RaftElectionTimeoutHigh - RaftElectionTimeoutLow)
		timeout := time.Duration(rand.Intn(floatInterval)) + RaftElectionTimeoutLow
		electTimer.Reset(timeout)
	}
}
