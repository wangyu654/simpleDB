package raft

func (rf *Raft) updateFollowerCommit(leaderCommit int, lastIndex int) {
	oldVal := rf.commitIndex
	if leaderCommit > rf.commitIndex {
		if leaderCommit < lastIndex {
			//不能超过leaderCommit
			rf.commitIndex = leaderCommit
		} else {
			//leaderCommit范围内的[:lastIndex]提交
			rf.commitIndex = lastIndex
		}
	}
	baseIndex := rf.log[0].Index
	for oldVal++; oldVal <= rf.commitIndex; oldVal++ {
		rf.chanCommitted <- ApplyMsg{
			CommandIndex: oldVal,
			Command:      rf.log[oldVal-baseIndex].Command,
			CommandValid: true,
		}
		rf.lastApplied = oldVal
	}
}

func (leader *Raft) updateLeaderCommit() {
	leader.mu.Lock()
	defer leader.mu.Unlock()
	defer leader.persist()

	// update commitIndex
	oldIndex := leader.commitIndex
	newIndex := oldIndex
	for i := len(leader.log) - 1; leader.log[i].Index > oldIndex && leader.log[i].Term == leader.currentTerm; i-- {
		countServer := 1
		for server := range leader.peers {
			if server == leader.me {
				continue
			}
			if leader.matchIndex[server] >= i {
				countServer++
			}
		}
		if countServer*2 >= len(leader.peers) {
			leader.printInfo("commited  log:", leader.log[i], ",received server count:", countServer)
			// 提前截断
			newIndex = i
			break
		}
	}
	if oldIndex == newIndex {
		return
	}
	leader.commitIndex = newIndex
	for i := oldIndex + 1; i <= newIndex; i++ {
		leader.chanCommitted <- ApplyMsg{
			CommandIndex: i,
			Command:      leader.log[i].Command,
			CommandValid: true,
		}
		leader.lastApplied = i
	}
}
