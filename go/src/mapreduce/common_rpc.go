package mapreduce

import (
	"fmt"
	"net/rpc"
)

// What follows are RPC types and methods.
// Field names must start with capital letters, otherwise RPC will break.

// DoTaskArgs holds the arguments that are passed to a worker when a job is scheduled on it.
// DoTaskArgs保存保护参数，用于为worker分配工作。
type DoTaskArgs struct {
	JobName    string   // 工作的名字
	File       string   // 待处理的文件名
	Phase      jobPhase // 工作类型是map还是reduce
	TaskNumber int      // 任务的索引？

	// NumOtherPhase is the total number of tasks in other phase; mappers
	// need this to compute the number of output bins, and reducers needs
	// this to know how many input files to collect.
	// 全部任务数量，mapper需要这个数字去计算输出的数量, 同时reducer需要知道有多少输入文件需要收集。
	NumOtherPhase int
}

// ShutdownReply is the response to a WorkerShutdown.
// It holds the number of tasks this worker has processed since it was started.
// ShutdownReply是WorkerShutdown的回应,Ntasks表示worker从启动开始已经处理的任务。
type ShutdownReply struct {
	Ntasks int
}

// RegisterArgs is the argument passed when a worker registers with the master.
// worker注册到master的时候，传递的参数。
type RegisterArgs struct {
	Worker string
}

// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be the address
// of a reply structure.
//
// call() returns true if the server responded, and false
// if call() was not able to contact the server. in particular,
// reply's contents are valid if and only if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in master.go, mapreduce.go,
// and worker.go.  please don't change this function.
//

// call（）使用参数args向服务器srv上的rpcname处理程序发送一个RPC，等待回复，然后回复答复。 回复参数应该是回复结构的地址。
//
// 如果服务器响应，则call（）返回true;如果call（）不能联系服务器，则返回false。 特别是，当且仅当call（）返回true时，回复的内容才有效。
//
// 你应该假设call（）将会超时并且在一段时间后返回一个错误，如果它没有得到来自服务器的回复的话。
//
//请使用call（）发送所有RPC，在master.go，mapreduce.go，和worker.go。 请不要更改此功能。
// 本地的rpc调用，使用是unix套接字
func call(srv string, rpcName string, args interface{}, reply interface{}) bool {
	c, err := rpc.Dial("unix", srv)
	if err != nil {
		return  false
	}
	defer c.Close()

	e := c.Call(rpcName,args,reply)
	if e == nil {
		 return  true
	}
	fmt.Println(err)
	return  false
}