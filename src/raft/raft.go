package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(Command interface{}) (index, Term, isleader)
//   start agreement on a new log entry
// rf.GetState() (Term, isLeader)
//   ask a Raft for its current Term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (

	// "fmt"
	// "log"
	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log Entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type State string

const (
	Follower  State = "follower"
	Candidate State = "candidate"
	Leader    State = "leader"
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	state State

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int
	votedFor    int
	log         []LogEntry
	lastReceive time.Time
}

type LogEntry struct {
	Command string
	Term    int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// Your code here (2A) //
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isleader = rf.state == Leader
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	//candidate’s Term
	Term int
	//candidate requesting vote
	CandidateId int
	//index of candidate’s last log entry (§5.4)
	//lastLogIndex int
	////Term of candidate’s last log entry (§5.4)
	//lastLogTerm int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	//currentTerm, for candidate to update itself
	Term int
	//true means candidate received vote
	VoteGranted bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	rf.mu.Lock()
	defer rf.mu.Unlock()
	// lock for this line
	reply.Term = rf.currentTerm
	Debug(dVote, "S%d requested by S%d, args.Term %d, rf.currentTerm %d, ", rf.me, args.CandidateId, args.Term, rf.currentTerm)
	//Reply false if Term < currentTerm (§5.1) -> too weak to be leader
	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		Debug(dVote, "S%d reject leader S%d", rf.me, args.CandidateId)
		return
	}

	if args.Term > rf.currentTerm{
		rf.convertToFollower(args.Term)
	}

	// 2. If votedFor is null or CandidateId, and candidate’s log is
	// at least as up-to-date as receiver’s log, grant vote (§5.2, §5.4)
	//&& args.lastLogIndex >= len(rf.log)
	reply.Term = rf.currentTerm
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		reply.VoteGranted = true
		Debug(dVote, "S%d grant leader for S%d", rf.me, args.CandidateId)
		return
	}
	Debug(dVote, "S%d WHAT THE HELL S%d", rf.me, args.CandidateId)
	return
}


// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next Command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// Command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the Command will appear at
// if it's ever committed. the second return value is the current
// Term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
	Debug(dError, "S%d killed", rf.me)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false{
		// Your code here (2A)
		// Check if a leader election should be started.
		ms := 360 + (rand.Int63() % 240)
		startTime := time.Now()
		// does we heard something, if yet, ignore send
		time.Sleep(time.Duration(ms) * time.Millisecond)
		if rf.lastReceive.Before(startTime) && rf.state != Leader {// does we heard something, if yet, ignore send
			rf.attemptElection()
		}
		// pause for a random amount of time between 50 and 350
		// milliseconds.
	}
}

func (rf *Raft) convertToCandidate() {
	rf.state = Candidate
	rf.currentTerm++
	rf.votedFor = rf.me
	rf.lastReceive = time.Now()
}

func (rf *Raft) convertToFollower(newTerm int) {
	rf.state = Follower
	rf.currentTerm = newTerm
	rf.votedFor = -1
	rf.lastReceive = time.Now()
}

func (rf *Raft) convertToLeader() {
	rf.state = Leader
	rf.lastReceive = time.Now()
}

func (rf *Raft) attemptElection() {
	rf.mu.Lock()
	rf.convertToCandidate()
	rf.mu.Unlock()
	//
	localTerm := rf.currentTerm
	count := 1
	done := false
	for anotherNode, _ := range rf.peers {
		if anotherNode == rf.me {
			continue
		}

		go func(atomicNode int) {
			vote, reply := rf.callRequestVote(atomicNode, localTerm)
			if !vote {
				return
			}
			rf.mu.Lock()
			defer rf.mu.Unlock()
			count++

			if count >= len(rf.peers)/2+1 {
				Debug(dLeader, "S%d become leader %d %d", rf.me, localTerm , reply.Term)
				rf.convertToLeader()
				go rf.pingFollower()
				done = true
			}
			if count == len(rf.peers) {
				done = true
			}
			if done && count < len(rf.peers)/2+1 {
				rf.convertToFollower(reply.Term)
				Debug(dLeader, "S%d not become leader", rf.me)
			}
		}(anotherNode)
	}
}

func (rf *Raft) callRequestVote(atomicNode int, localTerm int) (bool, *RequestVoteReply) {
	args := RequestVoteArgs{
		localTerm, // because it can be updated if it act as a receiver request.
		rf.me,
		//len(rf.log),
		//rf.log[len(rf.log)].Term,
	}
	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(atomicNode, &args, &reply)
	if !ok {
		//log.Printf("call failed!\n")
		return false, &reply
	}
	// reply.Y should be 100.
	//log.Printf("reply.IsNotify with Term %v and vote %v\n", reply.Term, reply.VoteGranted)
	return reply.VoteGranted, &reply
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {

	Debug(dTimer, "S%d here peers%+v, persi %+v", me, persister)

	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	// todo - check if this is correct
	rf.mu.Lock()
	rf.convertToFollower(0)
	rf.mu.Unlock()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}

func (rf *Raft) pingFollower() {
	for rf.state == Leader {
		ms := 100 + (rand.Int63() % 20)
		time.Sleep(time.Duration(ms) * time.Millisecond)

		for anotherNode, _ := range rf.peers {
			if anotherNode == rf.me {
				continue
			}
			go func(atomicNode int) {
				rf.callAppendEntries(atomicNode)
			}(anotherNode)
		}
	}
}

func (rf *Raft) callAppendEntries(atomicNode int) {
	rf.mu.Lock()
	args := AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderId:     rf.me,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      nil,
		LeaderCommit: 0,
	}
	reply := AppendEntriesReply{}
	Debug(dTimer, "S%d i;m leader, i'm ping %d", rf.me, atomicNode)
	rf.mu.Unlock()
	ok := rf.sendAppendEntries(atomicNode, &args, &reply)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if !ok {
		return
	}
	if reply.Success == false{
		Debug(dTimer, "S%d my term:%d ->follower with new term:%d", rf.me, rf.currentTerm, reply.Term)
		rf.convertToFollower(reply.Term)
	}
	return

}

// Term leader’s Term
// LeaderId so follower can redirect clients
// PrevLogIndex index of log entry immediately preceding new ones
// PrevLogTerm Term of PrevLogIndex entry
// Entries[] log Entries to store (empty for heartbeat; may send more than one for efficiency)
// LeaderCommit leader’s commitIndex
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}
type AppendEntriesReply struct {
	//Term currentTerm, for leader to update itself
	//Success true if follower contained entry matching PrevLogIndex and PrevLogTerm
	Term    int
	Success bool
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// Your code here (2A, 2B).
	//if toolong -> convert to candidate, auto leader selection
	// restart time = 0;
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//Debug(dInfo, "S%d current term:%d, caller %d with term:%d", rf.me,rf.currentTerm, args.LeaderId,args.Term )
	if args.Term < rf.currentTerm {
		reply.Success = false
		reply.Term = rf.currentTerm
		Debug(dTimer, "S%d reply S%d, you should be follower", rf.me, args.LeaderId)
		return
	}
	// holly shit
	reply.Success = true
	if args.Term > rf.currentTerm {
		Debug(dTimer, "S%d reply S%d, you are leader t:%d, i will be follower t:%d", rf.me, args.LeaderId, args.Term, rf.currentTerm)
		rf.convertToFollower(args.Term)
		return
	}
	return
}
