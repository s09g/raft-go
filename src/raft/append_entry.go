package raft

import "time"

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
	Conflict bool
	XTerm int
	XIndex int
	XLen int
}

func (rf *Raft) appendEntries(heartbeat bool) {
	lastLog := rf.lastLog()
	for peer, _ := range rf.peers {
		if peer == rf.me {
			rf.lastHeartBeat = time.Now()
			rf.resetElectionTimeout()
			continue
		}
		// rules for leader 3
		if lastLog.Index > rf.nextIndex[peer] || heartbeat {
			nextIndex := rf.nextIndex[peer]
			if nextIndex <= 0 {
				nextIndex = 1
			}
			if rf.lastLog().Index + 1 < nextIndex  {
				nextIndex = rf.lastLog().Index
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
	DPrintf("[%v]: term %v 发送entry to %v : args prev %v, lastLog %v", rf.me, args.Term, serverId, args.PrevLogIndex, args.PrevLogIndex + len(args.Entries))
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
			match := args.PrevLogIndex + len(args.Entries)
			next := match + 1
			rf.nextIndex[serverId] = max(rf.nextIndex[serverId], next)
			rf.matchIndex[serverId] = max(rf.matchIndex[serverId], match)
		} else
		if reply.Conflict {
			lastLogInXTerm := rf.findLastLogInTerm(reply.XTerm)
			DPrintf("[%v]: Conflict from %v %#v, lastLogInXTerm %v", rf.me, serverId, reply, lastLogInXTerm)
			if lastLogInXTerm > 0 {
				rf.nextIndex[serverId] = lastLogInXTerm
			} else {
				rf.nextIndex[serverId] = reply.XIndex
			}
			if reply.XLen < rf.nextIndex[serverId] {
				rf.nextIndex[serverId] = reply.XLen
			}
			DPrintf("[%v]: leader nextIndex[%v] %v log %v, ", rf.me, serverId, rf.nextIndex[serverId], rf.log)
		} else
		if rf.nextIndex[serverId] > 1 {
			rf.nextIndex[serverId]--
		}
		rf.leaderCommitRule()
	}
}

func (rf *Raft) findLastLogInTerm(x int) int {
	for i := len(rf.log) - 1; i > 0 ; i-- {
		term := rf.log[i].Term
		if term == x {
			return i
		} else if term < x {
			break
		}
	}
	return -1
}

func (rf *Raft) leaderCommitRule() {
	// leader rule 4
	if rf.state != Leader {
		return
	}
	N := rf.commitIndex
	for n := rf.commitIndex + 1; n <= rf.lastLog().Index; n++ {
		if rf.log[n].Term != rf.currentTerm {
			continue
		}
		counter := 1
		for serverId := 0; serverId < len(rf.peers); serverId++ {
			if serverId != rf.me && rf.matchIndex[serverId] >= n {
				counter++
			}
			if counter > len(rf.peers) / 2 {
				N = n
				break
			}
		}
	}
	if N == rf.commitIndex {
		return
	}
	rf.commitIndex = N
	DPrintf("[%v] leader尝试提交 index %v", rf.me, rf.commitIndex)
	rf.apply()
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("[%d]: follower 收到 [%v] AppendEntries %v, prevIndex %v, prevTerm %v, 当前log %v\n", rf.me, args.LeaderId, args.Entries, args.PrevLogIndex, args.PrevLogTerm, rf.log)
	// rules for servers
	// all servers 2
	reply.Success = false
	reply.Term = rf.currentTerm
	if args.Term > rf.currentTerm {
		rf.setNewTerm(args.Term)
	}

	// append entries rpc 1
	if args.Term < rf.currentTerm {
		return
	}
	//DPrintf("[%v]: reset heart beat", rf.me)
	rf.lastHeartBeat = time.Now()

	// candidate rule 3
	if rf.state == Candidate {
		rf.state = Follower
	}
	// append entries rpc 2
	if rf.lastLog().Index < args.PrevLogIndex || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.Conflict = true
		conflictIndex := min(args.PrevLogIndex, rf.lastLog().Index)
		xTerm := rf.log[conflictIndex].Term
		for xIndex := conflictIndex; xIndex > 0 ; xIndex-- {
			if rf.log[xIndex - 1].Term != xTerm {
				reply.XIndex = xIndex
				break
			}
		}
		reply.XTerm = xTerm
		reply.XLen = len(rf.log)
		DPrintf("[%v]: Conflict log %v", rf.me, rf.log)
		return
	}
	//DPrintf("[%v]: append entries rpc 2, log %v", rf.me, rf.log)

	// append entries rpc 3
	if args.PrevLogIndex < rf.lastLog().Index && rf.log[args.PrevLogIndex+1].Term != args.Term {
		rf.log = rf.log[: args.PrevLogIndex + 1]
		rf.persist()
	}
	//DPrintf("[%v]: append entries rpc 3, log %v", rf.me, rf.log)

	// append entries rpc 4
	for i, entry := range args.Entries {
		if entry.Index >= len(rf.log) || rf.log[entry.Index].Term != entry.Term {
			rf.log = append(rf.log, args.Entries[i:]...)
			rf.persist()
			break
		}
	}
	//DPrintf("[%v]: append entries rpc 4, log %v", rf.me, rf.log)

	// append entries rpc 5
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, rf.lastLog().Index)
		rf.apply()
	}
	reply.Success = true
	DPrintf("[%v]: follower commit %v, applied %v, 当前log %v", rf.me, rf.commitIndex, rf.lastApplied, rf.log)
}


func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


