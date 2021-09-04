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
	for i, _ := range rf.peers {
		if i == rf.me {
			rf.resetElectionTimeout()
			continue
		}
		if lastIndex > rf.nextIndex[i] || heartbeat {
			rf.sendEntries(i, heartbeat)
		}
	}
}


func (rf *Raft) sendEntries(peer int, heartbeat bool) {


}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
}


func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


