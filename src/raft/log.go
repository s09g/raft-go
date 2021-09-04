package raft

type Log struct {
	Command interface{}
	Index   int
	Term    int
}

func (rf *Raft) lastLog() *Log {
	return &rf.log[len(rf.log) - 1]
}

func (rf *Raft) appendLog(log *Log) {
	rf.log = append(rf.log, *log)
}

