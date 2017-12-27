package paxos

//
// Paxos library, to be included in an application.
// Multiple applications will run, each including
// a Paxos peer.
//
// Manages a sequence of agreed-on values.
// The set of peers is fixed.
// Copes with network failures (partition, msg loss, &c).
// Does not store anything persistently, so cannot handle crash+restart.
//
// The application interface:
//
// px = paxos.Make(peers []string, me int)
// px.Start(seq int, v interface{}) -- start agreement on new instance
// px.Status(seq int) (Fate, v interface{}) -- get info about an instance
// px.Done(seq int) -- ok to forget all instances <= seq
// px.Max() int -- highest instance seq known, or -1
// px.Min() int -- instances before this seq have been forgotten
//

// Paxos库，包含在应用程序中。多个应用程序将运行，每个应用程序都包含一个Paxos对等体。
//
// 管理一系列共识的值。
// 这组对等点是固定的。
// 处理网络故障（分区，msg丢失，＆c）。
// 不会持久存储任何东西，所以不能处理crash + restart。
//
// 应用程序 接口：
//
// px = paxos.Make（peers [] string，me int）
// px.Start（seq int，v interface {}） - 在新实例上启动协议
// px.Status（seq int）（Fate，v interface {}） - 获取有关实例的信息
// px.Done（seq int） - 可以放弃所有的实例<= seq
// px.Max（）int - 已知最高级实例seq，或-1
// px.Min（）int - 这个seq之前的实例已经被遗忘了

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
)

// px.Status() return values, indicating
// whether an agreement has been decided,
// or Paxos has not yet reached agreement,
// or it was agreed but forgotten (i.e. < Min()).
//
// px.Status（）返回值，表示协议是否已经决定，或者Paxos还没有达成一致，或者已经同意但忘记了（i.e. <Min（））。
type Fate int

const (
	Decided   Fate = iota + 1
	Pending        // not yet decided.尚未决定
	Forgotten      // decided but forgotten.决定但忘记了。
)

type Paxos struct {
	mu         sync.Mutex
	l          net.Listener
	dead       int32 // for testing
	unreliable int32 // for testing
	rpcCount   int32 // for testing
	peers      []string
	me         int // index into peers[]

	// Your data here.
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the replys contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it does not get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please do not change this function.
//
// call（）使用参数args向服务器srv上的rpcname处理程序发送一个RPC，等待回复，然后回复答复。 答复参数应该是一个指向答复结构的指针。
//
// 如果服务器响应，则返回值为true;如果call（）不能联系服务器，则返回false。 特别是，回复内容只有在call（）返回true时才有效。
//
// 你应该假设call（）会超时，如果它没有得到服务器的回复，会在一段时间后返回一个错误。
//
// 请使用call（）发送所有RPC，在client.go和server.go中。
// 请不要更改此功能。
//
func call(srv string, name string, args interface{}, reply interface{}) bool {
	c, err := rpc.Dial("unix", srv)
	if err != nil {
		err1 := err.(*net.OpError)
		if err1.Err != syscall.ENOENT && err1.Err != syscall.ECONNREFUSED {
			fmt.Printf("paxos Dial() failed: %v\n", err1)
		}
		return false
	}
	defer c.Close()

	err = c.Call(name, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

//
// the application wants paxos to start agreement on
// instance seq, with proposed value v.
// Start() returns right away; the application will
// call Status() to find out if/when agreement
// is reached.
//
// 应用程序希望paxos在实例seq上启动协议，建议值为v。
// Start（）立即返回; 应用程序将调用Status（）来查看是否/何时达成协议。
//
func (px *Paxos) Start(seq int, v interface{}) {
	// Your code here.
}

//
// the application on this machine is done with
// all instances <= seq.
//
// see the comments for Min() for more explanation.
//
//
// 这个机器上的应用程序是用所有的实例<= seq完成的。
// 请参阅Min（）的注释以获得更多解释。
//
func (px *Paxos) Done(seq int) {
	// Your code here.
}

//
// the application wants to know the
// highest instance sequence known to
// this peer.
//
//
//应用程序想要知道该对等方已知的最高实例序列。
//
func (px *Paxos) Max() int {
	// Your code here.
	return 0
}

//
// Min() should return one more than the minimum among z_i,
// where z_i is the highest number ever passed
// to Done() on peer i. A peers z_i is -1 if it has
// never called Done().
//
// Paxos is required to have forgotten all information
// about any instances it knows that are < Min().
// The point is to free up memory in long-running
// Paxos-based servers.
//
// Paxos peers need to exchange their highest Done()
// arguments in order to implement Min(). These
// exchanges can be piggybacked on ordinary Paxos
// agreement protocol messages, so it is OK if one
// peers Min does not reflect another Peers Done()
// until after the next instance is agreed to.
//
// The fact that Min() is defined as a minimum over
// *all* Paxos peers means that Min() cannot increase until
// all peers have been heard from. So if a peer is dead
// or unreachable, other peers Min()s will not increase
// even if all reachable peers call Done. The reason for
// this is that when the unreachable peer comes back to
// life, it will need to catch up on instances that it
// missed -- the other peers therefor cannot forget these
// instances.
//
//
// Min（）应该返回比z_i中最小的一个， 其中z_i是在对等体i上传递给Done（）的最高数字。如果从未调用Done（），则同位体z_i为-1。
//
// Paxos被要求忘记关于它知道的<Min（）的所有实例的所有信息。
// 关键是释放长时间运行的基于Paxos的服务器的内存。
//
// Paxos同伴需要交换最高的Done（）参数才能实现Min（）。
// 这些交换可以在普通的Paxos协议协议消息上捎带，所以如果一个对等者Min在下一个实例被同意之前不反映另一个Peers Done（），那么它是可以的。
//
// Min（）被定义为所有* Paxos对象的最小值的事实意味着Min（）不能增加，直到听到所有的同伴为止。
// 所以，如果一个对等体死亡或不可达，即使所有可达对等体都调用Done，其他对等体Min（）也不会增加。
// 原因是，当不可触及的对端恢复正常时，它需要赶上它所触发的实例 - 其他对等实体不能忘记这些实例。
//
func (px *Paxos) Min() int {
	// You code here.
	return 0
}

//
// the application wants to know whether this
// peer thinks an instance has been decided,
// and if so what the agreed value is. Status()
// should just inspect the local peer state;
// it should not contact other Paxos peers.
//
// 应用程序想知道这个同伴是否认为一个实例已经被确定，如果是的话，那么商定的值是什么。
// Status（）应该检查本地对等状态; 它不应该联系其他Paxos同行。
//
func (px *Paxos) Status(seq int) (Fate, interface{}) {
	// Your code here.
	return Pending, nil
}

//
// tell the peer to shut itself down.
// for testing.
// please do not change these two functions.
//
// 告诉同伴关闭自己。
// for testing.
// 请不要改变这两个功能。
//
func (px *Paxos) Kill() {
	atomic.StoreInt32(&px.dead, 1)
	if px.l != nil {
		px.l.Close()
	}
}

//
// has this peer been asked to shut down?
//
// 有这个断点peer被要求关闭？
func (px *Paxos) isdead() bool {
	return atomic.LoadInt32(&px.dead) != 0
}

// please do not change these two functions.
// 请不要更改这两个功能。
func (px *Paxos) setunreliable(what bool) {
	if what {
		atomic.StoreInt32(&px.unreliable, 1)
	} else {
		atomic.StoreInt32(&px.unreliable, 0)
	}
}

func (px *Paxos) isunreliable() bool {
	return atomic.LoadInt32(&px.unreliable) != 0
}

//
// the application wants to create a paxos peer.
// the ports of all the paxos peers (including this one)
// are in peers[]. this servers port is peers[me].
//
//
// 应用程序想要创建一个paxos对等体。
// 所有paxos对等端口（包括这个）
// 在peers[]中。 这个服务器端口是peers[me]。
func Make(peers []string, me int, rpcs *rpc.Server) *Paxos {
	px := &Paxos{}
	px.peers = peers
	px.me = me

	// Your initialization code here.

	if rpcs != nil {
		// caller will create socket &c
		rpcs.Register(px)
	} else {
		rpcs = rpc.NewServer()
		rpcs.Register(px)

		// prepare to receive connections from clients.
		// change "unix" to "tcp" to use over a network.
		os.Remove(peers[me]) // only needed for "unix"
		l, e := net.Listen("unix", peers[me])
		if e != nil {
			log.Fatal("listen error: ", e)
		}
		px.l = l

		// please do not change any of the following code,
		// or do anything to subvert it.
		// 请不要更改下面的任何代码，或者做任何事情来颠覆它。

		// create a thread to accept RPC connections
		// 创建一个线程来接受RPC连接
		go func() {
			for px.isdead() == false {
				conn, err := px.l.Accept()
				if err == nil && px.isdead() == false {
					if px.isunreliable() && (rand.Int63()%1000) < 100 {
						// discard the request.
						conn.Close()
					} else if px.isunreliable() && (rand.Int63()%1000) < 200 {
						// process the request but force discard of reply.
						c1 := conn.(*net.UnixConn)
						f, _ := c1.File()
						err := syscall.Shutdown(syscall.Handle(f.Fd()), syscall.SHUT_WR)
						if err != nil {
							fmt.Printf("shutdown: %v\n", err)
						}
						atomic.AddInt32(&px.rpcCount, 1)
						go rpcs.ServeConn(conn)
					} else {
						atomic.AddInt32(&px.rpcCount, 1)
						go rpcs.ServeConn(conn)
					}
				} else if err == nil {
					conn.Close()
				}
				if err != nil && px.isdead() == false {
					fmt.Printf("Paxos(%v) accept: %v\n", me, err.Error())
				}
			}
		}()
	}

	return px
}
