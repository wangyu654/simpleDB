package raft

type Role int

const (
	Follower  Role = 0
	Candidate Role = 1
	Leader    Role = 2
)

func (role Role) toString() string {
	if role == 0 {
		return "Follower"
	} else if role == 1 {
		return "Candidate"
	} else {
		return "Leader"
	}
}

func (rf *Raft) toFollower(votedID int) {
	rf.role = Follower
	rf.votedFor = -1
	rf.leaderId = votedID
	rf.chanRole <- Follower
}

func (rf *Raft) toLeader() {
	rf.role = Leader
	for i := range rf.peers {
		rf.nextIndex[i] = rf.log[len(rf.log)-1].Index + 1
		rf.matchIndex[i] = 0
	}
	rf.chanRole <- Leader
	rf.printInfo(rf.me, "becoming Leader")
}

func (rf *Raft) toCandidate() {
	rf.role = Candidate
	rf.chanRole <- Candidate
	rf.printInfo(rf.me, "becoming Candidate")
}
