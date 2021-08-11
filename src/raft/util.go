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

func (rf *Raft) getCurrentTerm() int {
	rf.mu.Lock()
	term := rf.currentTerm
	rf.mu.Unlock()
	return term
}

func (rf *Raft) setCurrentTerm(term int) {
	rf.mu.Lock()
	rf.currentTerm = term
	rf.mu.Unlock()
}

func (rf *Raft) incCurrentTerm()  {
	rf.mu.Lock()
	rf.currentTerm++
	rf.mu.Unlock()
}

func (rf *Raft) getVotedFor() int {
	rf.mu.Lock()
	votedFor := rf.votedFor
	rf.mu.Unlock()
	return votedFor
}

func (rf *Raft) setVotedFor(votedFor int) {
	rf.mu.Lock()
	rf.votedFor = votedFor
	rf.mu.Unlock()
}


func (rf *Raft) updateLastHeartBeat() {
	rf.mu.Lock()
	rf.lastHeatBeat = time.Now()
	rf.mu.Unlock()
}

func (rf *Raft) timeSinceLastHeartBeat() time.Duration {
	rf.mu.Lock()
	lastHeatBeat := rf.lastHeatBeat
	rf.mu.Unlock()
	return time.Since(lastHeatBeat)
}


func (rf *Raft) getRaftState() RaftState {
	rf.mu.Lock()
	state := rf.state
	rf.mu.Unlock()
	return state
}

func (rf *Raft) setRaftState(state RaftState) {
	rf.mu.Lock()
	rf.state = state
	rf.mu.Unlock()
}