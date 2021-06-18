package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"golab1/src/labgob"
	"golab1/src/labrpc"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// import "bytes"
// import "golab1/src/labgob"

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	log []LogEntry

	role        Role
	currentTerm int
	votedFor    int
	leaderId    int
	nextIndex   []int
	matchIndex  []int
	chanRole    chan Role

	lastApplied int
	commitIndex int

	// For election timer
	chanHeartbeat chan bool
	chanGrantVote chan bool
	chanCommitted chan ApplyMsg

	chanAppend []chan bool
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.currentTerm, rf.role == Leader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}

	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)

	rf.mu.Lock()
	d.Decode(&rf.currentTerm)
	d.Decode(&rf.votedFor)
	d.Decode(&rf.log)
	rf.mu.Unlock()

	DPrintf("[%d] read persist [%d]", rf.me, rf.log[len(rf.log)-1].Index)
}

//向leader提交cmd
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	// fmt.Println(rf.persister.mu)

	index := -1
	term := -1
	isLeader := false

	if rf.role == Leader {
		rf.mu.Lock()
		defer rf.mu.Unlock()
		defer rf.persist()

		rf.printInfo("receive log:", command)
		index = len(rf.log) + rf.log[0].Index
		term = rf.currentTerm
		isLeader = true
		newLog := LogEntry{}
		newLog.Command = command
		newLog.Term = term
		newLog.Index = index
		rf.log = append(rf.log, newLog)
	}

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	rf.currentTerm = 0
	rf.votedFor = -1
	rf.leaderId = -1
	rf.role = Follower
	rf.log = append(rf.log, LogEntry{Term: 0, Index: 0})
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))

	rf.chanHeartbeat = make(chan bool)
	rf.chanGrantVote = make(chan bool)
	rf.chanCommitted = applyCh
	rf.chanAppend = make([]chan bool, len(peers))
	for i := range rf.chanAppend {
		rf.chanAppend[i] = make(chan bool, 1)
		rf.chanAppend[i] <- true
	}

	rand.Seed(time.Now().UnixNano())

	rf.chanRole = make(chan Role)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.Run()
	go rf.startElectTimer()

	return rf
}

func (rf *Raft) Run() {
	role := rf.role
	for {
		switch role {
		case Leader:
			go rf.DoHeartbeat()
			role = <-rf.chanRole
		case Candidate:
			go rf.DoElection()
			role = <-rf.chanRole
		case Follower:
			role = <-rf.chanRole
		}
	}
}
