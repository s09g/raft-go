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
	DPrintf("[%v] 发送entry to %v : args index %v", rf.me, serverId, args.PrevLogIndex + len(args.Entries))
	var reply AppendEntriesReply
	ok := rf.sendAppendEntries(serverId, args, &reply)
	if !ok {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("[%v] %v reply append : reply %v", rf.me, serverId, reply)
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
		} else {
			if rf.nextIndex[serverId] > 1 {
				rf.nextIndex[serverId]--
			}
		}
	}

	rf.leaderCommitRule()
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
	DPrintf("[%d]: 收到 entry %v\n", rf.me, args)
	// rules for servers
	// all servers 2
	reply.Success = false
	reply.Term = rf.currentTerm
	if args.Term > rf.currentTerm {
		rf.setNewTerm(args.Term)
		return
	}

	// append entries rpc 1
	if args.Term < rf.currentTerm {
		return
	}

	// append entries rpc 2
	if len(rf.log) <= args.PrevLogIndex || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		return
	}

	// append entries rpc 3
	if args.PrevLogIndex + 1 > len(rf.log) && rf.log[args.PrevLogIndex+1].Term != args.Term {
		rf.log = rf.log[: args.PrevLogIndex + 1]
	}

	// append entries rpc 4
	for i, entry := range args.Entries {
		if entry.Index >= len(rf.log) || rf.log[entry.Index].Term != entry.Term {
			rf.log = append(rf.log, args.Entries[i:]...)
			break
		}
	}

	// append entries rpc 5
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, rf.lastLog().Index)
		DPrintf("[%d]: follower 确认commit log %v\n", rf.me, rf.commitIndex)
	}
	reply.Success = true
	rf.lastHeartBeat = time.Now()
	DPrintf("[%d]: 确认添加 log %v\n", rf.me, args.Entries)
	DPrintf("[%v] %#v", rf.me, rf)
}


func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


