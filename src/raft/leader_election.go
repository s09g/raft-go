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
	term := rf.CurrentTerm
	isleader := rf.state == Leader
	return term, isleader
}

func (rf *Raft) resetElectionTimeout() {
	rf.electionTimeout = time.Duration(150 + rand.Intn(150)) * time.Millisecond
}

func (rf *Raft) setNewTerm(term int) {
	if term > rf.CurrentTerm || rf.CurrentTerm == 0 {
		rf.state = Follower
		rf.CurrentTerm = term
		rf.VotedFor = -1
		DPrintf("[%d]: set term %v\n", rf.me, rf.CurrentTerm)
		rf.persist()
	}
}

func (rf *Raft) leaderElection() {
	rf.CurrentTerm++
	rf.state = Candidate
	rf.VotedFor = rf.me
	rf.persist()
	rf.lastHeartBeat = time.Now()
	rf.resetElectionTimeout()
	term := rf.CurrentTerm
	voteCounter := 1
	lastLog := rf.lastLog()
	DPrintf("[%v]: start leader election, term %d\n", rf.me, rf.CurrentTerm)
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
