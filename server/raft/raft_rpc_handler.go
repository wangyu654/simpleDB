package raft

// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	rf.printInfo("receive requestVote from", args.CandidateId)

	reply.VoteGranted = false
	reply.Term = rf.currentTerm

	if rf.currentTerm > args.Term {
		rf.printInfo("reject", args.CandidateId, ",term is older,", args.Term, "vs", rf.currentTerm)
		return
	}

	if rf.currentTerm == args.Term {
		if rf.votedFor != -1 && rf.votedFor != args.CandidateId {
			// 本轮已经投票
			rf.printInfo("reject ", args.CandidateId, "has voted")
			return
		}
	}

	if rf.currentTerm < args.Term {
		rf.currentTerm = args.Term
		reply.Term = args.Term
		rf.toFollower(args.CandidateId) //变为follwer
		if rf.log[len(rf.log)-1].Term > args.LastLogTerm || (rf.log[len(rf.log)-1].Term == args.LastLogTerm && rf.log[len(rf.log)-1].Index > args.LastLogIndex) {
			// 候选人的log更旧
			rf.printInfo("reject", args.CandidateId, ",log is older")
			return
		}
	}

	rf.vote(args, reply)
}

func (rf *Raft) vote(args RequestVoteArgs, reply *RequestVoteReply) {
	rf.chanGrantVote <- true
	reply.VoteGranted = true
	rf.printInfo("support", args.CandidateId)
}

func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	if args.Term < rf.currentTerm {
		rf.printInfo("reject hearbeat from leader, leader:", args.Term, "vs local:", rf.currentTerm)
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	}

	rf.chanHeartbeat <- true

	rf.printInfo("receive hearbeat from leader, leader:", args.Term, "vs local:", rf.currentTerm)
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		if rf.role != Follower {
			rf.toFollower(args.LeaderId)
		}
	}
	reply.Term = args.Term
	baseLogIndex := rf.log[0].Index

	if args.PrevLogIndex < baseLogIndex {
		reply.Success = false
		reply.NextIndex = baseLogIndex + 1
		return
	}

	if args.PrevLogIndex >= baseLogIndex+len(rf.log) {
		//要比较的index大于peer的日志范围
		reply.Success = false
		reply.NextIndex = baseLogIndex + len(rf.log)
		return
	}

	if rf.log[args.PrevLogIndex-baseLogIndex].Term != args.PrevLogTerm {
		//preTerm不匹配
		reply.Success = false
		nextIndex := args.PrevLogIndex
		failTerm := rf.log[args.PrevLogIndex-baseLogIndex].Term
		for nextIndex >= baseLogIndex && rf.log[nextIndex-baseLogIndex].Term == failTerm {
			//跨过一个term
			nextIndex--
		}
		reply.NextIndex = nextIndex
		return
	}
	var updateLogIndex int
	if len(args.Entry) != 0 {
		//非心跳 且preIndex的term匹配上了  updatelog
		// rf.printInfo("receive entries", args.Entry)
		for i, entry := range args.Entry {
			curIndex := entry.Index - baseLogIndex //leader传播的日志日志的插入位置
			if curIndex < len(rf.log) {
				if entry.Term != rf.log[curIndex].Term || entry.Command != rf.log[curIndex].Command {
					//curIndex 已经存在日志 强势leader 直接删除该处日志
					rf.log = append(rf.log[:curIndex], entry)
				} //else 已经存在这个entry
			} else {
				//直接续上即可
				rf.log = append(rf.log, args.Entry[i:]...)
				break
			}
		}
		reply.NextIndex = rf.log[len(rf.log)-1].Index + 1 //peer下一个要匹配的位置
		updateLogIndex = rf.log[len(rf.log)-1].Index      //peer 经过leader同步后的最新日志的Index
	} else {
		// reply.
		// if args.LeaderIsEmpty { //! special leader刚收到cmd就被0日志的节点推翻，此时边界条件比较复杂！！
		// 	rf.log = []LogEntry{{Term: 0, Index: 0}}
		// }
		reply.NextIndex = args.PrevLogIndex + 1 //第一次 PrevLogIndex=0
		updateLogIndex = args.PrevLogIndex
	}
	reply.Success = true
	rf.updateFollowerCommit(args.LeaderCommit, updateLogIndex)
}
