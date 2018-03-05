package raft_x

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//
// 这是Raft必须暴露给服务（或测试者）的API的概要。 有关更多详细信息，请参阅下面的每个函数的注释。

// rf = Make(...) // create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader) // start agreement on a new log entry
// rf.GetState() (term, isLeader) // 问一个raft的当前任期，以及它是否认为是领导

// ApplyMsg
// 每当一个新条目被提交到日志时，每个Raft对等方都应该向同一个服务器中的服务（或测试者）发送一个ApplyMsg。

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"labrpc"
	"math/rand"
	"sync"
	"time"
)

// 定义节点的raft状态
const (
	STATE_LEADER = iota
	STATE_CANDIDATE
	STATE_FLLOWER

	HBINTERVAL = 50 * time.Millisecond // 50ms 心跳时间
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
//
//由于每个Raft对等方都知道提交了连续的日志条目，所以对等方应该通过传递给Make（）的applyCh将ApplyMsg发送到同一服务器上的服务（或测试者）。
//
type ApplyMsg struct {
	Index   int         // 日志索引的ID
	TermId  int         // Raft选举的ID
	Command interface{} // 相关命令接口

	UseSnapshot bool   // 是否使用快照版本
	Snapshot    []byte // 快照信息
}

type LogEntry struct {
	LogIndex  int // 日志索引的ID
	LogTermId int
	LogComd   interface{}
}

//
// A Go object implementing a single Raft peer.
//
// Go对象实现一个简单的Raft节点实例
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd // 节点
	persister *Persister          // 持久化日志逻辑类
	me        int                 // 当前节点在所有节点的位置

	//查看论文的图2中的描述-Raft服务的状态必须维护。
	state         int //节点状态（Leader，Candidate，Follower）三种之一
	voteCount     int //已知的最大的已经被提交的日志条目的索引值
	chanCommit    chan bool
	chanHearBeat  chan bool
	chanGrantVote chan bool
	chanLeader    chan bool
	chanApply     chan ApplyMsg

	// 持久化状态信息
	currentTermId int        //服务器最后一次知道的任期号（初始化为 0，持续递增），需要持久化
	votedFor      int        //在当前term获得选票的候选人 Id，需要持久化
	log           []LogEntry //日志条目集；每一个条目包含一个用户状态机执行的指令，和收到时的任期号，需要持久化

	// 对所有节点可见的状态信息
	commitIndex int //已知的最大的已经被提交的日志条目的索引值
	lastApplied int //最后被应用到状态机的日志条目索引值（初始化为 0，持续递增）

	// volatile state on leader
	nextIndex  []int //对于每一个服务器，需要发送给他的下一个日志条目的索引值（初始化为领导人最后索引值加一），leader有用
	matchIndex []int //对于每一个服务器，已经复制给他的日志的最高索引值，leader有用
}

// return currentTerm and whether this server
// believes it is the leader.
//返回currentTerm以及这个服务器是否认为它是领导者。
func (raft *Raft) GetState() (int, bool) {
	return raft.currentTermId, raft.state == STATE_LEADER
}
func (raft *Raft) GetLastIndex() int {
	return raft.log[len(raft.log)-1].LogIndex
}
func (raft *Raft) GetLastTermId() int {
	return raft.log[len(raft.log)-1].LogTermId
}
func (raft *Raft) IsLeader() bool {
	return raft.state == STATE_LEADER
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
//将Raft的持久状态保存到稳定的存储中， 在崩溃并重新启动后，它可以在以后被检索到。
//请参阅图2的描述，以了解什么应该是持久的。
func (raft *Raft) persist() {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	encoder.Encode(raft.currentTermId)
	encoder.Encode(raft.votedFor)
	encoder.Encode(raft.log)
	data := buffer.Bytes()

	raft.persister.SaveRaftState(data)
}

func (raft *Raft) readSnapshot(data []byte) {

	if len(data) == 0 {
		return
	}
	raft.readPersist(raft.persister.ReadRaftState())

	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	var LastIncludedIndex int
	var LastIncludedTermId int
	decoder.Decode(LastIncludedIndex)
	decoder.Decode(LastIncludedTermId)

	raft.commitIndex = LastIncludedIndex
	raft.currentTermId = LastIncludedTermId

	raft.log = truncateLog(LastIncludedIndex, LastIncludedTermId, raft.log)

	msg := ApplyMsg{
		UseSnapshot: true,
		Snapshot:    data,
	}
	// TODO
	go func() {
		raft.chanApply <- msg
	}()
}

//
// restore previously persisted state.
// 恢复以前存储的状态。
func (raft *Raft) readPersist(data []byte) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&raft.currentTermId)
	decoder.Decode(&raft.votedFor)
	decoder.Decode(&raft.log)
}

// 转换持久化的日志信息。
func truncateLog(lastIncludedIndex int, lastIncludedTermId int, log []LogEntry) []LogEntry {
	var newLogEntries []LogEntry
	newLogEntries = append(newLogEntries, LogEntry{
		LogIndex:  lastIncludedIndex,
		LogTermId: lastIncludedTermId,
	})
	for index := len(log) - 1; index >= 0; index-- {
		if log[index].LogIndex == lastIncludedIndex && log[index].LogTermId == lastIncludedTermId {
			newLogEntries = append(newLogEntries, log[index+1:]...)
			break
		}
	}
	return newLogEntries
}

// ========================= TODO ================================

// 发起RPC投票的请求参数
type RequestVoteArgs struct {
	TermId        int // 投票ID
	CandidateId   int // 候选人ID
	LastLogTermId int
	LastLogIndex  int
}

// 发起RPC投票的响应参数
type RequestVoteReply struct {
	TermId      int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	TermId        int        //领导人的任期号
	LeaderId      int        //领导人的 Id，以便于跟随者重定向请求
	PrevLogTermId int        //prevLogIndex条目的任期值
	PrevLogIndex  int        //新的日志条目紧随之前的索引值
	Entries       []LogEntry //需要存储当然日志条目（表示心跳时为空；一次性发送多个是为了提高效率）
	LeaderCommit  int        //领导人已经提交的日志的索引值
}
type AppendEntriesReply struct {
	TermId    int  //当前的任期号，用于领导人去更新自己
	Success   bool //跟随者包含了匹配上 prevLogIndex 和 prevLogTerm 的日志时为真
	NextIndex int  // TODO
}

//=====================================================
// TODO
func (raft *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	raft.mu.Lock()
	defer raft.mu.Unlock()
	defer raft.persist()

	reply.VoteGranted = false
	if args.TermId < raft.currentTermId {
		reply.TermId = raft.currentTermId
		return
	}
	if args.TermId > raft.currentTermId {
		raft.currentTermId = args.TermId
		raft.state = STATE_FLLOWER
		raft.votedFor = -1
	}
	reply.TermId = raft.currentTermId

	termId := raft.GetLastTermId()
	index := raft.GetLastIndex()
	uptoDate := false

	if args.LastLogTermId > termId {
		uptoDate = true
	}
	if args.LastLogTermId == termId && args.LastLogIndex >= index {
		uptoDate = true
	}
	if (raft.votedFor == -1 || raft.votedFor == args.CandidateId) && uptoDate {
		raft.chanGrantVote <- true
		raft.state = STATE_FLLOWER
		reply.VoteGranted = true
		raft.votedFor = args.CandidateId
		fmt.Printf("%v currentTerm:%v vote for:%v term:%v", raft.me, raft.currentTermId, args.CandidateId, args.TermId)
	}
}

// TODO
func (raft *Raft) AppendEntries(args AppendEntriesArgs, reply AppendEntriesReply) {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	defer raft.persist()
	reply.Success = false
	if args.TermId < raft.currentTermId {
		reply.TermId = raft.currentTermId
		reply.NextIndex = raft.GetLastIndex() + 1
		return
	}
	raft.chanHearBeat <- true
	if args.TermId > raft.currentTermId {
		raft.currentTermId = args.TermId
		raft.state = STATE_FLLOWER
		raft.votedFor = -1
	}
	reply.TermId = args.TermId

	if args.PrevLogIndex > raft.GetLastIndex() {
		reply.NextIndex = raft.GetLastIndex() + 1
		return
	}

	baseIndex := raft.log[0].LogIndex
	if args.PrevLogIndex > baseIndex {
		termId := raft.log[args.PrevLogIndex-baseIndex].LogTermId
		if args.PrevLogTermId != termId {
			for i := args.PrevLogIndex - 1; i >= baseIndex; i-- {
				if raft.log[i-baseIndex].LogTermId != termId {
					reply.NextIndex = i + 1
					break
				}
			}
			return
		}
	} /*else {
		//fmt.Printf("????? len:%v\n",len(args.Entries))
		last := rf.getLastIndex()
		elen := len(args.Entries)
		for i := 0; i < elen ;i++ {
			if args.PrevLogIndex + i > last || rf.log[args.PrevLogIndex + i].LogTerm != args.Entries[i].LogTerm {
				rf.log = rf.log[: args.PrevLogIndex+1]
				rf.log = append(rf.log, args.Entries...)
				app = false
				fmt.Printf("?????\n")
				break
			}
		}
	}*/
	if args.PrevLogIndex < baseIndex {

	} else {
		raft.log = raft.log[:args.PrevLogIndex+1-baseIndex]
		raft.log = append(raft.log, args.Entries...)
		reply.Success = true
		reply.NextIndex = raft.GetLastIndex() + 1
	}
	println(raft.me, raft.GetLastIndex(), reply.NextIndex, raft.log)
	if args.LeaderCommit > raft.commitIndex {
		last := raft.GetLastIndex()
		if args.LeaderCommit > last {
			raft.commitIndex = last
		} else {
			raft.commitIndex = args.LeaderCommit
		}
		raft.chanCommit <- true
	}
	return
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
// 将RequestVote RPC发送到服务器的示例代码。
// server是rf.peers []中目标服务器的索引。
// 期望args中的RPC参数。
// 用RPC回复填充*reply，所以调用者应该通过并回复。
// 传递给Call（）的参数和回复的类型必须与处理函数中声明的参数类型相同（包括它们是否是指针）。
// 如果labrpc说RPC已经交付，则返回true。
// 如果您无法使RPC正常工作，请检查您是否已经对通过RPC传递的结构中的所有字段名进行了大写处理，
// 并且调用者使用＆而不是结构本身传递了答复结构的地址。
func (raft *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := raft.peers[server].Call("Raft.RequestVote", args, reply)
	raft.mu.Lock()
	defer raft.mu.Unlock()

	if ok {
		termId := raft.currentTermId
		if raft.state != STATE_CANDIDATE {
			return ok
		}
		if args.TermId != termId {
			return ok
		}
		if reply.TermId > termId {
			raft.currentTermId = reply.TermId
			raft.state = STATE_FLLOWER
			raft.votedFor = -1
			raft.persist()
		}
		if reply.VoteGranted {
			raft.voteCount++
			if raft.state == STATE_CANDIDATE && raft.voteCount > len(raft.peers)/2 {
				raft.state = STATE_FLLOWER
				raft.chanLeader <- true
			}
		}
	}
	return ok
}

// 发送日志追加逻辑
func (raft *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := raft.peers[server].Call("Raft.AppendEntries", args, reply)
	raft.mu.Lock()
	defer raft.mu.Unlock()

	if ok {
		if raft.state != STATE_LEADER {
			return ok
		}
		if args.TermId != raft.currentTermId {
			return ok
		}

		if reply.TermId > raft.currentTermId {
			raft.currentTermId = reply.TermId
			raft.state = STATE_FLLOWER
			raft.votedFor = -1
			raft.persist()
			return ok
		}
		if reply.Success {
			if len(args.Entries) > 0 {
				raft.nextIndex[server] = args.Entries[len(args.Entries)-1].LogIndex + 1
				raft.matchIndex[server] = raft.nextIndex[server] - 1
			}
		} else {
			raft.nextIndex[server] = reply.NextIndex
		}
	}
	return ok
}

type InstallSnapshotArgs struct {
	TermId             int
	LeaderId           int
	LastIncludedIndex  int
	LastIncludedTermId int
	Data               []byte
}

type InstallSnapshotReply struct {
	TermId int
}

func (raft *Raft) GetPerisistSize() int {
	return raft.persister.RaftStateSize()
}

func (raft *Raft) StartSnapshot(snapshot []byte, index int) {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	baseIndex := raft.log[0].LogIndex
	lastIndex := raft.GetLastIndex()

	if index <= baseIndex || index > lastIndex {
		// in case having installed a snapshot from leader before snapshotting
		// second condition is a hack
		//在快照之前安装了来自leader的快照
		//第二个条件是hack
		return
	}
	var newLogEntries []LogEntry

	newLogEntries = append(newLogEntries, LogEntry{
		LogIndex:  index,
		LogTermId: raft.log[index-baseIndex].LogTermId,
	})

	for i := index + 1; i <= lastIndex; i++ {
		newLogEntries = append(newLogEntries, raft.log[i-baseIndex])
	}

	raft.log = newLogEntries

	raft.persist()

	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	encoder.Encode(newLogEntries[0].LogIndex)
	encoder.Encode(newLogEntries[0].LogTermId)

	data := buffer.Bytes()
	data = append(data, snapshot...)

	raft.persister.SaveRaftSnapshot(data)
}

func (raft *Raft) InstallSnapshot(args InstallSnapshotArgs, reply *InstallSnapshotReply) {
	// Your code here.
	raft.mu.Lock()
	defer raft.mu.Unlock()
	if args.TermId < raft.currentTermId {
		reply.TermId = raft.currentTermId
		return
	}
	raft.chanHearBeat <- true
	raft.state = STATE_FLLOWER
	raft.currentTermId = reply.TermId

	raft.persister.SaveRaftSnapshot(args.Data)

	raft.log = truncateLog(args.LastIncludedIndex, args.LastIncludedTermId, raft.log)

	msg := ApplyMsg{UseSnapshot: true, Snapshot: args.Data}

	raft.lastApplied = args.LastIncludedIndex
	raft.commitIndex = args.LastIncludedIndex

	raft.persist()

	raft.chanApply <- msg
}

func (raft *Raft) sendInstallSnapshot(server int, args InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := raft.peers[server].Call("Raft.InstallSnapshot", args, reply)
	if ok {
		if reply.TermId > raft.currentTermId {
			raft.currentTermId = reply.TermId
			raft.state = STATE_FLLOWER
			raft.votedFor = -1
			return ok
		}

		raft.nextIndex[server] = args.LastIncludedIndex + 1
		raft.matchIndex[server] = args.LastIncludedIndex
	}
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
//
// 使用Raft的服务（例如k / v服务器）想要开始将下一个命令的协议附加到Raft的日志中。
// 如果此服务器不是领导者，则返回false。 否则立即启动协议并返回。 这不能保证
// 因为领导者可能会失败或失去一次选举，所以指挥官将永远致力于Raft日志。
//
// 第一个返回值是该命令在提交时将出现的索引。
// 第二个返回值是当前项。 如果此服务器认为它是领导者，则第三个返回值为true。
func (raft *Raft) Start(command interface{}) (int, int, bool) {
	raft.mu.Lock()
	defer raft.mu.Unlock()
	index := -1
	termId := raft.currentTermId
	isLeader := raft.state == STATE_LEADER

	if isLeader {
		index = raft.GetLastIndex() + 1
		fmt.Printf("raft:%d start\n", raft.me)
		raft.log = append(raft.log, LogEntry{
			LogTermId: termId,
			LogIndex:  index,
			LogComd:   command,
		})
		raft.persist()
	}
	return index, termId, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
//
// 当Raft实例不再需要时，测试者调用Kill（）。
// 您不需要在Kill（）中执行任何操作，但可能会很方便（例如）关闭此实例的调试输出。
//

func (raft *Raft) kill() {

}

// 广播通知投票信息
func (raft *Raft) boatcastRequestVote() {
	var args RequestVoteArgs
	raft.mu.Lock()
	args.TermId = raft.currentTermId
	args.CandidateId = raft.me
	args.LastLogIndex = raft.GetLastIndex()
	args.LastLogTermId = raft.GetLastTermId()
	raft.mu.Unlock()

	for i := range raft.peers {
		if i != raft.me && raft.state == STATE_CANDIDATE {
			go func(i int) {
				var reply RequestVoteReply
				fmt.Printf("%v RequestVote to %v\n", raft.me, i)
				raft.sendRequestVote(i, args, &reply)
			}(i)
		}
	}
}

func (raft *Raft) boatcastAppendEntries() {
	raft.mu.Lock()

	defer raft.mu.Unlock()
	N := raft.commitIndex
	last := raft.GetLastIndex()
	baseIndex := raft.log[0].LogIndex
	for i := raft.commitIndex + 1; i <= last; i++ {
		num := 1
		for j := range raft.peers {
			if j != raft.me && raft.matchIndex[j] >= i && raft.log[i-baseIndex].LogTermId == raft.currentTermId {
				num++
			}
		}
		if 2*num > len(raft.peers) {
			N = i
		}
	}
	if N != raft.commitIndex {
		raft.commitIndex = N
		raft.chanCommit <- true
	}

	for i := range raft.peers {
		if i != raft.me && raft.state == STATE_LEADER {

			//copy(args.Entries, rf.log[args.PrevLogIndex + 1:])

			if raft.nextIndex[i] > baseIndex {
				var args AppendEntriesArgs
				args.TermId = raft.currentTermId
				args.LeaderId = raft.me
				args.PrevLogIndex = raft.nextIndex[i] - 1
				//	fmt.Printf("baseIndex:%d PrevLogIndex:%d\n",baseIndex,args.PrevLogIndex )
				args.PrevLogTermId = raft.log[args.PrevLogIndex-baseIndex].LogTermId
				//args.Entries = make([]LogEntry, len(rf.log[args.PrevLogIndex + 1:]))
				args.Entries = make([]LogEntry, len(raft.log[args.PrevLogIndex+1-baseIndex:]))
				copy(args.Entries, raft.log[args.PrevLogIndex+1-baseIndex:])
				args.LeaderCommit = raft.commitIndex
				go func(i int, args AppendEntriesArgs) {
					var reply AppendEntriesReply
					raft.sendAppendEntries(i, args, &reply)
				}(i, args)
			} else {
				var args InstallSnapshotArgs
				args.TermId = raft.currentTermId
				args.LeaderId = raft.me
				args.LastIncludedIndex = raft.log[0].LogIndex
				args.LastIncludedTermId = raft.log[0].LogTermId
				args.Data = raft.persister.snapshot
				go func(server int, args InstallSnapshotArgs) {
					reply := &InstallSnapshotReply{}
					raft.sendInstallSnapshot(server, args, reply)
				}(i, args)
			}
		}
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
// 服务或测试者想要创建一个Raft服务器。所有的Raft服务器（包括这个）的端口都是保存在peers[]中。
// 这台服务器的端口是peers[me]。 所有服务器的peers []数组具有相同的顺序。
// persister是这个服务器保存其持久状态的地方，并且最初保存最近保存的状态（如果有的话）。
// applyCh是测试人员或服务期望Raft发送ApplyMsg消息的通道。 Make（）必须快速返回，所以它应该为任何长时间运行的工作启动goroutines。
func CreateRaftServer(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.persister = persister
	rf.peers = peers
	rf.me = me

	rf.state = STATE_FLLOWER
	rf.votedFor = -1
	rf.log = append(rf.log, LogEntry{LogTermId: 0})
	rf.currentTermId = 0
	rf.chanCommit = make(chan bool, 100)
	rf.chanHearBeat = make(chan bool, 100)

	return nil
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here.
	rf.state = STATE_FLLOWER
	rf.votedFor = -1
	rf.log = append(rf.log, LogEntry{LogTermId: 0})
	rf.currentTermId = 0
	rf.chanCommit = make(chan bool, 100)
	rf.chanHearBeat = make(chan bool, 100)
	rf.chanGrantVote = make(chan bool, 100)
	rf.chanLeader = make(chan bool, 100)
	rf.chanApply = applyCh

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	rf.readSnapshot(persister.ReadSnapshot())

	go func() {
		for {
			switch rf.state {
			case STATE_FLLOWER:
				select {
				case <-rf.chanHearBeat:
				case <-rf.chanGrantVote:
				case <-time.After(time.Duration(rand.Int63()%333+550) * time.Millisecond):
					rf.state = STATE_CANDIDATE
				}
			case STATE_LEADER:
				//fmt.Printf("Leader:%v %v\n",rf.me,"boatcastAppendEntries	")
				rf.boatcastAppendEntries()
				time.Sleep(HBINTERVAL)
			case STATE_CANDIDATE:
				rf.mu.Lock()
				rf.currentTermId++
				rf.votedFor = rf.me
				rf.voteCount = 1
				rf.persist()
				rf.mu.Unlock()
				go rf.boatcastRequestVote()
				//fmt.Printf("%v become CANDIDATE %v\n",rf.me,rf.currentTerm)
				select {
				case <-time.After(time.Duration(rand.Int63()%333+550) * time.Millisecond):
				case <-rf.chanHearBeat:
					rf.state = STATE_FLLOWER
					//	fmt.Printf("CANDIDATE %v reveive chanHeartbeat\n",rf.me)
				case <-rf.chanLeader:
					rf.mu.Lock()
					rf.state = STATE_LEADER
					//fmt.Printf("%v is Leader\n",rf.me)
					rf.nextIndex = make([]int, len(rf.peers))
					rf.matchIndex = make([]int, len(rf.peers))
					for i := range rf.peers {
						rf.nextIndex[i] = rf.GetLastIndex() + 1
						rf.matchIndex[i] = 0
					}
					rf.mu.Unlock()
					//rf.boatcastAppendEntries()
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-rf.chanCommit:
				rf.mu.Lock()
				commitIndex := rf.commitIndex
				baseIndex := rf.log[0].LogIndex
				for i := rf.lastApplied + 1; i <= commitIndex; i++ {
					msg := ApplyMsg{Index: i, Command: rf.log[i-baseIndex].LogComd}
					applyCh <- msg
					//fmt.Printf("me:%d %v\n",rf.me,msg)
					rf.lastApplied = i
				}
				rf.mu.Unlock()
			}
		}
	}()
	return rf
}
