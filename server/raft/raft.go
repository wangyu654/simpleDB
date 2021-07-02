package raft

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"
)

type Raft struct {
	mu        sync.Mutex    // Lock to protect shared access to this peer's state
	peers     []*rpc.Client // RPC end points of all peers
	loseConn  []bool        // the connection state of client
	addresses []string
	persister *Persister // Object to hold this peer's persisted state
	me        int        // this peer's index into peers[]
	dead      int32      // set by Kill()

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

func (rf *Raft) GetState() (int, bool) {

	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.currentTerm, rf.role == Leader
}

func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

func (rf *Raft) readPersist(data []byte) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}

	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)

	d.Decode(&rf.currentTerm)
	d.Decode(&rf.votedFor)
	d.Decode(&rf.log)

	DPrintf("[%d] read persist [%d]", rf.me, rf.log[len(rf.log)-1].Index)
}

//向leader提交cmd
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()
	index := -1
	term := -1
	isLeader := false

	if rf.role == Leader {

		// rf.printInfo("receive log:", command)
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

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func Make(peers []*rpc.Client, me int,
	persister *Persister, applyCh chan ApplyMsg, addresses []string) *Raft {
	rf := new(Raft)
	rpc.Register(rf)
	rf.peers = peers
	rf.loseConn = make([]bool, len(peers))
	rf.addresses = addresses
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

	rf.readPersist(persister.ReadRaftState(me))

	go rf.ChangeRole()
	go rf.startElectTimer()
	go rf.Sysmon()

	log.Println("raft", rf.me, "start")
	return rf
}

func (rf *Raft) ChangeRole() {
	role := rf.role
	for {
		// fmt.Println(rf.chanRole)
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
