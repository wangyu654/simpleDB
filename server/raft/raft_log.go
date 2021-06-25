package raft

type LogEntry struct {
	Command interface{}
	Term    int
	Index   int
}
