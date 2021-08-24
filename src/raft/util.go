package raft

import (
	"log"
	"time"
)

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func (rf *Raft) timeSinceLastHeartBeat() time.Duration {
	rf.mu.Lock()
	lastHeatBeat := rf.lastHeatBeat
	rf.mu.Unlock()
	return time.Since(lastHeatBeat)
}

func (rf *Raft) setNewTerm(term int) {
	rf.currentTerm = term
	rf.state = Follower
	rf.votedFor = -1
}