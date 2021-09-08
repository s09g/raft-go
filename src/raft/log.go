package raft

type Log struct {
	Command interface{}
	Term    int
	Index   int
}

func (rf *Raft) lastLog() *Log {
	return &rf.Logs[len(rf.Logs) - 1]
}

