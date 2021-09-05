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
	lastIndex := rf.lastLog().Index
	for peer, _ := range rf.peers {
		if peer == rf.me {
			rf.resetElectionTimeout()
			continue
		}
		// rules for leader 3
		if lastIndex > rf.nextIndex[peer] || heartbeat {
			nextIndex := rf.nextIndex[peer]
			lastLog := rf.lastLog()
			if nextIndex <= 0{
				nextIndex = 1
			}

			if nextIndex > lastLog.Index+ 1 {
				nextIndex = lastLog.Index
			}
			prevLog := rf.log[nextIndex - 1]
			args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				PrevLogIndex: prevLog.Index,
				PrevLogTerm:  prevLog.Term,
				Entries:      make([]Log, lastLog.Index - nextIndex + 1),
				LeaderCommit: rf.commitIndex,
			}
			copy(args.Entries, rf.log[nextIndex:])
			go rf.leaderSendEntries(peer, &args)
		}
	}
}

func (rf *Raft) leaderSendEntries(serverId int, args *AppendEntriesArgs) {
	var reply AppendEntriesReply
	ok := rf.sendAppendEntries(serverId, args, &reply)
	if !ok {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if reply.Term > rf.currentTerm {
		rf.setNewTerm(reply.Term)
		return
	}
	if reply.Term == rf.currentTerm {
		// rules for leader 3.1
		if reply.Success {
			rf.nextIndex[serverId]++
			rf.matchIndex[serverId]++
		} else {
			rf.nextIndex[serverId]--
		}
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// rules for servers
	// all servers 2
	if args.Term > rf.currentTerm {
		rf.setNewTerm(args.Term)
	}
	reply.Term = rf.currentTerm
	// append entries rpc 1
	if args.Term < rf.currentTerm {
		reply.Success = false
		return
	}
	// append entries rpc 2
	if len(rf.log) <= args.PrevLogIndex || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.Success = false
		return
	}
	// append entries rpc 3 & 4
	if len(rf.log) - 1 >= args.PrevLogIndex {
		// append entries rpc 3
		rf.log = rf.log[: args.PrevLogIndex + 1]
		// append entries rpc 4
		rf.log = append(rf.log, args.Entries...)
	}
	// append entries rpc 5
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, rf.lastLog().Index)
	}
	reply.Success = true
}


func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


