package raft

import (
	"math/rand"
	"sync"
	"time"
)

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term := rf.currentTerm
	isleader := rf.state == Leader
	return term, isleader
}

func (rf *Raft) resetElectionTimer() {
	t := time.Now()
	electionTimeout := time.Duration(150 + rand.Intn(150)) * time.Millisecond
	rf.electionTime = t.Add(electionTimeout)
}

func (rf *Raft) setNewTerm(term int) {
	if term > rf.currentTerm || rf.currentTerm == 0 {
		rf.state = Follower
		rf.currentTerm = term
		rf.votedFor = -1
		DPrintf("[%d]: set term %v\n", rf.me, rf.currentTerm)
		rf.persist()
	}
}

func (rf *Raft) leaderElection() {
	rf.currentTerm++
	rf.state = Candidate
	rf.votedFor = rf.me
	rf.persist()
	rf.resetElectionTimer()
	term := rf.currentTerm
	voteCounter := 1
	lastLog := rf.lastLog()
	DPrintf("[%v]: start leader election, term %d\n", rf.me, rf.currentTerm)
	args := RequestVoteArgs{
		Term:         term,
		CandidateId:  rf.me,
		LastLogIndex: lastLog.Index,
		LastLogTerm:  lastLog.Term,
	}

	var becameLeader sync.Once
	for serverId, _ := range rf.peers {
		if serverId != rf.me {
			go rf.candidateRequestVote(serverId, &args, &voteCounter, &becameLeader)
		}
	}
}
