package raft

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
//	每当一个新条目被提交到日志时，每个Raft对等方都应该向同一个服务器中的服务（或测试者）发送一个ApplyMsg。


import (
	"bytes"
	"encoding/gob"
	"labrpc"
	"math/rand"
	"sync"
	"time"
)

type raft struct {

}