package raft

type AppendEntriesArgs struct {
	Term int
	LeaderId int
	PrevLogIndex int
	PrevLogTerm int
	Entries []Log
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term int
	Success bool
}

func (rf *Raft) appendEntries(heartbeat bool) {
	lastIndex := rf.lastLog().index
	for peer, _ := range rf.peers {
		if peer == rf.me {
			rf.resetElectionTimeout()
			continue
		}
		if lastIndex > rf.nextIndex[peer] || heartbeat {
			nextIndex := rf.nextIndex[peer]
			lastLog := rf.lastLog()
			if nextIndex <= 0{
				nextIndex = 1
			}

			if nextIndex > lastLog.index + 1 {
				nextIndex = lastLog.index
			}
			prevLog := rf.log[nextIndex - 1]
			args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				PrevLogIndex: prevLog.index,
				PrevLogTerm:  prevLog.term,
				Entries:      make([]Log, lastLog.index - nextIndex + 1),
				LeaderCommit: rf.commitIndex,
			}
			copy(args.Entries, rf.log[nextIndex:])
			go rf.leaderSendEntries(peer, &args)
		}
	}
}



func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


