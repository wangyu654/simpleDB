package raft

import (
	"fmt"
	"log"
	"strconv"
)

// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

func (rf *Raft) printInfo(strings ...interface{}) {
	fmt.Println(rf.PrefixString(), strings)
}

func (rf *Raft) PrefixString() string {
	return "[" + "id " + strconv.Itoa(rf.me) + " " + rf.role.toString() + " term " + strconv.Itoa(rf.currentTerm) + " votedFor " + strconv.Itoa(rf.votedFor) + "]"
}
