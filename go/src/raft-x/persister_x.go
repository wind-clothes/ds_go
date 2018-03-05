package raft_x

import (
	"sync"
)

//
// support for Raft and kvraft to save persistent
// Raft state (log &c) and k/v server snapshots.
//
// we will use the original persister.go to test your code for grading.
// so, while you can modify this code to help you debug, please
// test with the original before submitting.
//
//
//支持Raft和kvraft以保存持久的Raft状态（log ＆c）和k / v服务器快照。
//
//我们将使用原始的persister.go来测试你的代码进行评分。
//所以，虽然你可以修改这个代码来帮助你调试，但是在提交之前请用原来的测试代码。
//

type Persister struct {
	mu         sync.Mutex // 互斥锁
	raft_state []byte     // raft的状态
	snapshot   []byte     // 快照信息 TODO
}

func MakePersister() *Persister {
	return &Persister{}
}

// 拷贝raft持久数据
func (ps *Persister) Copy() *Persister {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	np := MakePersister()
	np.raft_state = ps.raft_state
	np.snapshot = ps.snapshot
	return np
}
func (ps *Persister) SaveRaft(data []byte, snapShot []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.raft_state = data
	ps.snapshot = snapShot
}

// 保存raft持久数据
func (ps *Persister) SaveRaftState(data []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.raft_state = data
}

// 读取raft持久数据的状态信息
func (ps *Persister) ReadRaftState() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.raft_state
}

//保存raft持久数据快照数据
func (ps *Persister) SaveRaftSnapshot(data []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.snapshot = data
}

// 读取raft持久数据的快照信息
func (ps *Persister) ReadSnapshot() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.snapshot
}

func (ps *Persister) RaftStateSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return len(ps.raft_state)
}