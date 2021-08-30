package raft

import "time"

type AppendEntryArgs struct {
	Term        int
	LeaderID int
	PrevLogIndex int
	PrevLogTerm int
	Entries []Entry
	leaderCommit int
}

type AppendEntryReply struct {
	Term int
	Success bool
}

func (rf *Raft) appendEntries(index int) {
	if index < 0 {

	}
	rf.mu.Lock()
	term := rf.currentTerm
	rf.mu.Unlock()
	args := AppendEntryArgs{
		Term:        term,
		LeaderID: rf.me,
	}
	reply := AppendEntryReply{}
	DPrintf("[%d] leader state %#v", rf.me, rf.state)
	failures := 1
	finished := true

	for serverId, _ := range rf.peers {
		if serverId == rf.me {
			rf.mu.Lock()
			rf.lastHeatBeat = time.Now()
			rf.mu.Unlock()
			continue
		}
		go func(serverId int) {
			ack := rf.sendEntry(serverId, &args, &reply)
			if !ack {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				failures++
				if finished || failures <= len(rf.peers)/2 {
					DPrintf("[%d] 失联个数 %d\n", rf.me, failures)
					return
				}
				finished = true
				rf.state = Follower
			}
		}(serverId)
	}

}

func (rf *Raft) AppendEntry(args *AppendEntryArgs, reply *AppendEntryReply) {
	DPrintf("[%d]: 收到 %d 心跳 对方term %d\n", rf.me, args.LeaderID, args.Term)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.currentTerm {
		return
	}
	DPrintf("[%d]: 收到 %d 心跳 当前 term %d state %v\n", rf.me, args.LeaderID, rf.currentTerm, rf.state)
	rf.setNewTerm(args.Term)
	rf.lastHeatBeat = time.Now()
	DPrintf("[%d]: 收到 %d 心跳 最终 term %d state %v\n", rf.me, args.LeaderID, rf.currentTerm, rf.state)
}

func (rf *Raft) sendEntry(server int, args *AppendEntryArgs, reply *AppendEntryReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntry", args, reply)
	return ok
}
