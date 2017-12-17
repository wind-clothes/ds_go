// Package mapreduce provides a simple mapreduce library with a sequential
// implementation. Applications should normally call Distributed() [located in
// master.go] to start a job, but may instead call Sequential() [also in
// master.go] to get a sequential execution for debugging purposes.
//
// The flow of the mapreduce implementation is as follows:
//
//   1. The application provides a number of input files, a map function, a
//      reduce function, and the number of reduce tasks (nReduce).
//   2. A master is created with this knowledge. It spins up an RPC server (see
//      master_rpc.go), and waits for workers to register (using the RPC call
//      Register() [defined in master.go]). As tasks become available (in steps
//      4 and 5), schedule() [schedule.go] decides how to assign those tasks to
//      workers, and how to handle worker failures.
//   3. The master considers each input file one map tasks, and makes a call to
//      doMap() [common_map.go] at least once for each task. It does so either
//      directly (when using Sequential()) or by issuing the DoJob RPC on a
//      worker [worker.go]. Each call to doMap() reads the appropriate file,
//      calls the map function on that file's contents, and produces nReduce
//      files for each map file. Thus, there will be #files x nReduce files
//      after all map tasks are done:
//
//          f0-0, ..., f0-0, f0-<nReduce-1>, ...,
//          f<#files-1>-0, ... f<#files-1>-<nReduce-1>.
//
//   4. The master next makes a call to doReduce() [common_reduce.go] at least
//      once for each reduce task. As for doMap(), it does so either directly or
//      through a worker. doReduce() collects nReduce reduce files from each
//      map (f-*-<reduce>), and runs the reduce function on those files. This
//      produces nReduce result files.
//   5. The master calls mr.merge() [master_splitmerge.go], which merges all
//      the nReduce files produced by the previous step into a single output.
//   6. The master sends a Shutdown RPC to each of its workers, and then shuts
//      down its own RPC server.
//
// TODO:
// You will have to write/modify doMap, doReduce, and schedule yourself. These
// are located in common_map.go, common_reduce.go, and schedule.go
// respectively. You will also have to write the map and reduce functions in
// ../main/wc.go.
//
// You should not need to modify any other files, but reading them might be
// useful in order to understand how the other methods fit into the overall
// architecture of the system.

// 包mapreduce提供了一个简单的mapreduce库和一个串行执行顺序的实现。应用程序通常应该调用Distributed() [位于 master.go]来启动一个job，
// 但是也可以调用Sequential（）[也在master.go中]获得一个顺序执行的调试目的。
//
// mapreduce实现的流程如下：
//
// 1.该应用程序提供了许多输入文件，一个map函数，一个reduce函数和一些reduce tasks（nReduce）。
// 2.创建这个任务的master。它启动一个RPC服务器（见master_rpc.go），并等待worker注册（使用RPC调用Register（）[在master.go中定义]）。
//   当任务变得可用时（在步骤4和5中），schedule（）[schedule.go]决定如何将这些任务分配给worker，以及如何处理worker失败。
// 3.master将每个输入文件视为一个map tasks，并为每个任务至少调用一次doMap（）[common_map.go]。它直接执行（当使用Sequential（））
//   或通过在worker [worker.go]上发出DoJob RPC时。每次调用doMap（）都会读取相应的文件，调用该文件内容的map函数，并为每个map文件生成nReduce文件。因此，
//   完成所有映射任务后将会有#files x nReduce文件：
// 		f0-0，...，f0-0，f0- <nReduce-1>，...，
// 		f <＃files-1> -0，... f <＃files-1> - <nReduce-1>。
//
// 4.接下来master为每个reduce tasks调用doReduce（）[common_reduce.go]至少一次。正如doMap（），
//   它直接或者是通过一名work执行。 doReduce（）从每个收集nReduce reduce文件 map（f - * - <reduce>），
//   并在这些文件上运行reduce函数。这会产生nReduce结果文件。
// 5.master调用mr.merge（）[master_splitmerge.go]，它将所有的合并在一起 上一步产生的nReduce文件转换为单个输出。
// 6.master发送一个Shutdown RPC给它的每个worker，然后关闭下自己的RPC服务器。
//
// TODO：
// 您将不得不编写/修改doMap，doReduce方法，并自己安排。这些分别位于common_map.go，common_reduce.go和schedule.go。
// 你也必须在编写map并reduce功能在../main/wc.go。
//
//你不需要修改任何其他文件，但是可能会读取它们 有助于理解其他方法如何融入整体系统的体系结构。
package mapreduce
