package raft

type Log struct {
	entries []Entry
	offset int
}

func (l *Log) lastIndex() int {
	index := l.offset + len(l.entries) - 1
	return index
}

func (l *Log) append(e *Entry) {
	l.entries = append(l.entries, *e)
}

type Entry struct {
	Command interface{}
	Index int
	Term int
}