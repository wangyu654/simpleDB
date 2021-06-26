package raft

import (
	"fmt"
	"log"
	"time"
)

func (leader *Raft) DoHeartbeat() {
	for index := range leader.peers {
		if index != leader.me {
			go func(i int) { //每个peer 起一个协程发送心跳
				heartbeatTimer := time.NewTimer(RaftHeartbeatPeriod)
				for leader.role == Leader {
					leader.doAppendEntries(i)
					heartbeatTimer.Reset(RaftHeartbeatPeriod)
					<-heartbeatTimer.C
				}
			}(index)
			continue
		}
		go func() {
			heartbeatTimer := time.NewTimer(RaftHeartbeatPeriod)
			for leader.role == Leader {
				leader.chanHeartbeat <- true // 更新chanHeartbeat计时器
				leader.updateLeaderCommit()
				heartbeatTimer.Reset(RaftHeartbeatPeriod)
				<-heartbeatTimer.C
			}
		}()
	}
}

func (leader *Raft) doAppendEntries(server int) {

	/*

		+--------------------+
			保证向同一个peer
			不重复发送很多rpc
		+--------------------+

	*/
	select {
	case <-leader.chanAppend[server]:
	default:
		return
	}
	defer func() {
		leader.chanAppend[server] <- true
	}()

	// 准备rpc args
	leader.mu.Lock()
	baseLogIndex := leader.log[0].Index
	args := AppendEntriesArgs{}
	args.LeaderId = leader.me
	args.Term = leader.currentTerm
	args.LeaderCommit = leader.commitIndex
	if leader.nextIndex[server] > baseLogIndex { //有日志需要复制
		args.PrevLogIndex = leader.nextIndex[server] - 1
		args.PrevLogTerm = leader.log[args.PrevLogIndex].Term
		if leader.nextIndex[server] < leader.log[0].Index+len(leader.log) {
			args.Entry = leader.log[leader.nextIndex[server]-leader.log[0].Index:]
		}
	}
	leader.printInfo("send heartbeat to", server, "logs:", args.Entry)
	if len(leader.log) == 1 {
		args.LeaderIsEmpty = true
	}
	leader.mu.Unlock()
	reply := &AppendEntriesReply{}
	if leader.sendAppendEntries(server, args, reply) {
		leader.mu.Lock()
		if leader.role != Leader {
			leader.mu.Unlock()
			return
		}
		if args.Term != leader.currentTerm || reply.Term > args.Term {
			if reply.Term > args.Term {
				leader.currentTerm = reply.Term
				leader.toFollower(server)
				leader.persist()
			}
			leader.mu.Unlock()
			return
		}
		if reply.Success {
			leader.matchIndex[server] = reply.NextIndex - 1
			leader.nextIndex[server] = reply.NextIndex
		} else {
			leader.nextIndex[server] = reply.NextIndex
		}
		leader.mu.Unlock()
	}
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	var ret bool
	c := make(chan bool)
	go func() {
		rf.printInfo("send heartbeat to ", server)
		err := rf.peers[server].Call("Raft.AppendEntries", args, reply)
		if err == nil {
			c <- true
		} else {
			log.Println("failed to send heart because of", err)
		}
	}()
	select {
	case ret = <-c:
		// case <-time.After(RaftHeartbeatPeriod):
		// 	rf.printInfo("send heartbeat failed to ", server, "because of net")
		// 	ret = false
	}
	return ret
}

func (candidate *Raft) DoElection() {
	//初始化选举rpc参数
	candidate.mu.Lock()
	candidate.votedFor = candidate.me
	candidate.currentTerm++
	args := RequestVoteArgs{}
	args.CandidateId = candidate.me
	args.Term = candidate.currentTerm
	args.LastLogIndex = candidate.log[len(candidate.log)-1].Index
	args.LastLogTerm = candidate.log[len(candidate.log)-1].Term
	candidate.mu.Unlock()
	chanGather := make(chan bool, len(candidate.peers)) //缓冲chan
	chanGather <- true                                  //自投
	for index := range candidate.peers {
		if index == candidate.me {
			continue
		}
		go func(index int) {
			reply := &RequestVoteReply{}
			// candidate.printInfo("request vote to", index)
			if candidate.sendRequestVote(index, args, reply) {
				//接收到了消息
				fmt.Printf("receive requestVote reply:+%v", reply)
				if reply.VoteGranted {
					chanGather <- true
				} else {
					candidate.mu.Lock()
					candidate.printInfo(index, "reject me")
					if candidate.currentTerm < reply.Term {
						candidate.currentTerm = reply.Term
					}
					candidate.persist()
					candidate.mu.Unlock()
					chanGather <- false
				}
			} else {
				chanGather <- false
			}
		}(index)
		candidate.mu.Lock()
		if args.Term != candidate.currentTerm {
			return
		}
		candidate.mu.Unlock()
	}
	candidate.handleElectionRes(chanGather)
}

func (candidate *Raft) handleElectionRes(chanGather chan bool) {
	yes, no, countTotal := 0, 0, 0
	isLoop := true
	for isLoop {
		select {
		case ok := <-chanGather:
			countTotal++
			if ok {
				yes++
			} else {
				no++
			}
			if yes*2 > len(candidate.peers) {
				candidate.mu.Lock()
				candidate.toLeader()
				candidate.persist()
				candidate.printInfo("vote count:", yes)
				candidate.mu.Unlock()
				isLoop = false
			}
			if no*2 > len(candidate.peers) {
				isLoop = false
			}
		}
	}
}

func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	var ret bool
	c := make(chan bool)
	go func() {
		rf.printInfo("send reqeustVote to ", server)
		err := rf.peers[server].Call("Raft.RequestVote", args, reply)
		if err == nil {
			c <- true
		} else {
			log.Println("failed to send heart because of", err)
		}
	}()
	select {
	case ret = <-c:
		// case <-time.After(RaftHeartbeatPeriod):
		// 	//未接收消息
		// 	rf.printInfo("send reqeustVote failed to ", server, "because of net")
		// 	ret = false
	}
	return ret
}
