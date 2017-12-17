package mapreduce

import (
	"fmt"
	"net"
	"sync"
)

// Master holds all the state that the master needs to keep track of. Of
// particular importance is registerChannel, the channel that notifies the
// master of workers that have gone idle and are in need of new work.
//master有master需要跟踪的所有状态。 RegisterChannel是一个特别重要的渠道，通知那些空闲的workers，需要新的work。
type Master struct {
	mu sync.Mutex

	address         string      // master 的地址
	registerChannel chan string //与worker是否是空闲的传输通道
	doneChannel     chan bool
	workers         []string // thread safe for mutex

	jobName   string   //当前执行任务的名称
	files     []string // input files
	toNReduce int      // Number of reduce 分区

	shutdown chan interface{} // 关闭worker的通道
	l        net.Listener     // 网络监听
	stats    []int            // 状态集合
}

// newMaster initializes a new Map/Reduce Master
// newMaster 是初始化Map/Reduce Master 实例
func newMaster(master string) (mr *Master) {
	mr = &Master{
		address:         master,
		shutdown:        make(chan interface{}),
		registerChannel: make(chan string),
		doneChannel:     make(chan bool),
	}
	return mr
}

// Register is an RPC method that is called by workers after they have started
// up to report that they are ready to receive tasks.
// Register是RPC方法，在开始报告准备好接收任务之后由worker调用。
func (mr *Master) Register(rWorkers *RegisterArgs, d *interface{}) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	debug("Register: worker %s\n", rWorkers.Worker)

	mr.workers = append(mr.workers, rWorkers.Worker)
	go func() {
		mr.registerChannel <- rWorkers.Worker
	}()
	return nil
}

// Sequential runs map and reduce tasks sequentially, waiting for each task to
// complete before scheduling the next.
// 串行顺序运行按顺序map 和 reduce tasks，等待每个task完成后再调度下一个task。
func Sequential(jobName string, files []string, toNReduce int,
	mapF func(string, string) []KeyValue, reduceF func(string, []string) string, master string) (mr *Master) {
	mr = newMaster(master)

	go mr.run(jobName, files, toNReduce, func(phase jobPhase) {
		switch phase {
		case mapPhase:
			for i, f := range mr.files {
				doMap(mr.jobName, i, f, mr.toNReduce, mapF)
			}
		case reducePhase:
			for i := 0; i < mr.toNReduce; i++ {
				doReduce(mr.jobName, i, len(mr.files), reduceF)
			}
		}
	}, func() {
		mr.stats = []int{len(files) + toNReduce}
	})
	return
}

// Distributed schedules map and reduce tasks on workers that register with the master over RPC.
// 分布式的调度通过RPC向master注册的worker执行map和reduce tasks
func Distributed(jobName string, files []string, toNReduce int, master string) (mr *Master) {
	mr = newMaster(master)
	mr.startRpcServer()
	go mr.run(jobName, files, toNReduce, mr.schedule, func() {
		mr.stats = mr.killWorkers()
		mr.stopRpcServer()
	})
	return
}

// run executes a mapreduce job on the given number of mappers and reducers.
//
// First, it divides up the input file among the given number of mappers, and
// schedules each task on workers as they become available. Each map task bins
// its output in a number of bins equal to the given number of reduce tasks.
// Once all the mappers have finished, workers are assigned reduce tasks.
//
// When all tasks have been completed, the reducer outputs are merged,
// statistics are collected, and the master is shut down.
//
// Note that this implementation assumes a shared file system.

// 在指定的mapper和reducer数量上面执行mapreduce工作.
// 首先,在指定数量的mapper上面分配输入文件，然后分配每个任务到可用的worker。每个map任务将它的输出
// 放置在一些“箱子”, 数量等于给定的reduce任务的数量。一旦全部的mapper工作完成，worker开始安排reduce任务。
//
// 当全部的任务完成的时候,reducer的输出被合并,统计被收集，然后master关闭退出。
//
// 注意：实现假设在一个共享的文件系统之上。
func (mr *Master) run(jobName string, files []string, toNReduce int, schedule func(phase jobPhase), finish func()) {

	mr.jobName = jobName
	mr.files = files
	mr.toNReduce = toNReduce

	fmt.Printf("%s: Starting Map/Reduce task %s\n", mr.address, mr.jobName)

	schedule(mapPhase)    // 执行doMap
	schedule(reducePhase) // 执行doReduce
	finish()

	mr.merge()

	fmt.Printf("%s: Map/Reduce task completed\n", mr.address)
	mr.doneChannel <- true

}

// Wait blocks until the currently scheduled work has completed.
// This happens when all tasks have scheduled and completed, the final output
// have been computed, and all workers have been shut down.
// Wait 块，直到当前scheduled的work完成。
// 当所有task已经scheduled和完成，最终的输出已经被计算出来，并且所有的worker都被关闭时，执行等待。
func (mr *Master) Wait() {
	<-mr.doneChannel
}

// killWorkers cleans up all workers by sending each one a Shutdown RPC.
// It also collects and returns the number of tasks each worker has performed.
// killWorkers通过发送每个关闭RPC清除所有worker。
// 它还收集并返回每个worker执行的task数量。
func (mr *Master) killWorkers() []int {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	ntasks := make([]int, 0, len(mr.workers))

	for _, w := range mr.workers {
		debug("Master: shutdown worker %s\n", w)
		reply := &ShutdownReply{}
		ok := call(w, "Worker.Shutdown", new(interface{}), reply)
		if ok {
			ntasks = append(ntasks, reply.Ntasks)
		} else {
			fmt.Printf("Master: RPC %s shutdown error\n", w)
		}

	}
	return ntasks
}
