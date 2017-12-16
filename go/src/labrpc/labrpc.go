package labrpc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"
)

//
// channel-based RPC, for 824 labs.
// allows tests to disconnect RPC connections.
//
// we will use the original labrpc.go to test your code for grading.
// so, while you can modify this code to help you debug, please
// test against the original before submitting.
//
// adapted from Go net/rpc/server.go.
//
// sends gob-encoded values to ensure that RPCs
// don't include references to program objects.
//
// net := MakeNetwork() -- holds network, clients, servers.
// end := net.MakeEnd(endname) -- create a client end-point, to talk to one server.
// net.AddServer(servername, server) -- adds a named server to network.
// net.DeleteServer(servername) -- eliminate the named server.
// net.Connect(endname, servername) -- connect a client to a server.
// net.Enable(endname, enabled) -- enable/disable a client.
// net.Reliable(bool) -- false means drop/delay messages
//
// end.Call("Raft.AppendEntries", &args, &reply) -- send an RPC, wait for reply.
// the "Raft" is the name of the server struct to be called.
// the "AppendEntries" is the name of the method to be called.
// Call() returns true to indicate that the server executed the request
// and the reply is valid.
// Call() returns false if the network lost the request or reply
// or the server is down.
// It is OK to have multiple Call()s in progress at the same time on the
// same ClientEnd.
// Concurrent calls to Call() may be delivered to the server out of order,
// since the network may re-order messages.
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return. That is, there
// is no need to implement your own timeouts around Call().
// the server RPC handler function must declare its args and reply arguments
// as pointers, so that their types exactly match the types of the arguments
// to Call().
//
// srv := MakeServer()
// srv.AddService(svc) -- a server can have multiple services, e.g. Raft and k/v
//   pass srv to net.AddServer()
//
// svc := MakeService(receiverObject) -- obj's methods will handle RPCs
//   much like Go's rpcs.Register()
//   pass svc to srv.AddService()
//

// RPC回复消息的数据结构
type replyMsg struct {
	ok    bool   // 请求是否成功
	reply []byte //响应的信息字节流
}

// RPC请求消息
type reqMsg struct {
	endName  interface{}   // name of sening clientend 接口
	svcMeth  string        // 参数类型
	argsType reflect.Type  // 参数类型
	args     []byte        // 请求参数
	replyCh  chan replyMsg // 接收回复的通道
}

// =============================================//
//  创建网络                                     //
//  rn := MakeNetwork()                         //
//  创建客户端,并存在于此网络                     //
//   e := rn.MakeEnd("end1-99")                 //
//  创建服务，服务提供者是JunkServer              //
//   js := &JunkServer{}                        //
//   svc := MakeService(js)                     //
//  创建服务器对象，并添加到此网络                 //
//    rs := MakeServer()                        //
//    rs.AddService(svc)                        //
//    rn.AddServer("server99", rs)              //
//  客户端连接到服务器                           //
//     rn.Connect("end1-99", "server99")        //
//    rn.Enable("end1-99", true)                //
//==============================================//

// 创建网络连接的模型：记录一些网络上面的连接信息
type Network struct {
	mu             sync.Mutex                  // 同步锁
	reliable       bool                        // 标记网络是否可靠
	longDelays     bool                        // pause a long time on send on disabled connection
	longReordering bool                        // 随机的延迟回复
	ends           map[interface{}]*ClientEnd  //ends, by name
	enabled        map[interface{}]bool        // 判断是否还在存活在这个此网络上
	servers        map[interface{}]*Server     // servers, by name （管理这个网络上面的全部服务器程序）
	connections    map[interface{}]interface{} // endname -> servername（客户端和服务器端的连接关系）
	endCh          chan reqMsg
}

// 创造模拟网络对象-创建网络
func MakeNetwork() *Network {
	rn := &Network{
		reliable:    true,                         //默认网络是可靠的，不可靠只是在发生回应的时候随机延迟一段时间和随机放弃请求包
		ends:        map[interface{}]*ClientEnd{}, //存在于这个网络的端点
		enabled:     map[interface{}]bool{},       //
		servers:     map[interface{}]*Server{},
		connections: map[interface{}]interface{}{}, // 客户端和服务端之间的连接关系
		endCh:       make(chan reqMsg),
	}
	go func() {
		for xreq := range rn.endCh {
			go rn.ProcessRequest(xreq)
		}
	}()
	return rn
}

// 处理ClientEnd.Call(),每一个Call都在一个单独的goroutine中处理
func (rn *Network) ProcessRequest(req reqMsg) {
	enabled, servername, server, reliable, longreordering := rn.ReadEndnameInfo(req.endName)

	// 开始发送请求
	if enabled && servername != nil && server != nil {
		if reliable == false {
			// 模拟延迟
			ms := (rand.Int() % 27)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
		if reliable == false && (rand.Int()%1000) < 100 {
			// drop the request, return as if timeout
			req.replyCh <- replyMsg{false, nil}
			return
		}
		// 执行请求（RPC call handler），在一个单独的线程当中方便后期我们的check
		// 如果服务器返回失败则RPC需要返回一个失败的回复
		ech := make(chan replyMsg)
		go func() {
			r := server.dispatch(req) // 分发请求
			ech <- r
		}()
		//等待处理程序返回， 但如果调用了DeleteServer（），则不要再等待， 并返回一个错误。
		var reply replyMsg
		replyOK, serverDead := false, false
		for replyOK == false && serverDead == false {
			select {
			case reply = <-ech: // 等待server.dispatch返回结果
				replyOK = true
			case <-time.After(100 * time.Millisecond): // 等待超时
				serverDead = rn.IsServerDead(req.endName, servername, server)
			}
		}
		// 如果DeleteServer()被调用就不响应回复，比如服务器已经被杀死。
		// 这是为了避免这样的一个情况：客户端获得对Append的肯定回复，但服务器持久更新进入老persister。 ？？？
		// config.go会很小心的在取代Persister之前调用DeleteServer()。
		serverDead = rn.IsServerDead(req.endName, servername, server)
		if replyOK == false || serverDead == true {
			// server was killed while we were waiting; return error.
			req.replyCh <- replyMsg{false, nil}
		} else if reliable == false && (rand.Int()%1000) < 100 {
			// drop the reply, return as if timeout
			req.replyCh <- replyMsg{false, nil}
		} else if longreordering == true && rand.Intn(900) < 600 {
			// delay the response for a while
			ms := 200 + rand.Intn(1+rand.Intn(2000))
			time.Sleep(time.Duration(ms) * time.Millisecond)
			req.replyCh <- reply
		} else {
			req.replyCh <- reply
		}
	} else {
		// 模拟没有回复和最终超时。
		ms := 0
		if rn.longDelays {
			// let Raft tests check that leader doesn't send
			// RPCs synchronously.
			ms = (rand.Int() % 7000)
		} else {
			// many kv tests require the client to try each
			// server in fairly rapid succession.
			ms = (rand.Int() % 100)
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		req.replyCh <- replyMsg{false, nil}
	}
}

// 通过检索Network中的网络信息确定名字为endname的服务器是否还活着
func (rn *Network) IsServerDead(endName interface{}, servername interface{}, server *Server) bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.enabled[endName] == false || rn.servers[servername] != server {
		return true
	}
	return false
}

// 通过名字获取到配置信息
func (rn *Network) ReadEndnameInfo(endname interface{}) (enabled bool, servername interface{},
	server *Server, reliable bool, longreordering bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	enabled = rn.enabled[endname]        // 获取访问服务是否可用
	servername = rn.connections[endname] // 获取服务的服务器地址
	if servername != nil {
		server = rn.servers[servername] // 根据服务器名称获取Server实例
	}
	reliable = rn.reliable
	longreordering = rn.longReordering
	return
}

// connect a ClientEnd to a server.
// a ClientEnd can only be connected once in its lifetime.
// 客户端连接到服务器，客户端在自己的一个生命周期内只能连接一次
func (rn *Network) Connect(endname interface{}, servername interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.connections[endname] = servername // key：客户端，value：服务器端
}

// enable/disable a ClientEnd.
func (rn *Network) Enable(endname interface{}, enabled bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.enabled[endname] = enabled // 使客户端存活或者消失在这个此网络上
}

func (rn *Network) Reliable(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.reliable = yes
}
func (rn *Network) LongReordering(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.longReordering = yes // sometimes delay replies a long time
}
func (rn *Network) LongDelays(yes bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.longDelays = yes // pause a long time on send on disabled connection
}

// 通过名字获取到Server，然后获取到rpc请求数量
func (rn *Network) GetCount(servername interface{}) int {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	svr, ok := rn.servers[servername]
	if ok {
		return svr.GetCount()
	}
	return 0
}

// ===============================server=====================================
// server是一系列服务器的集合，全部服务共享这同一个rpc分发器（rpc dispatcher）。
// 这样Raft和键值服务都可以监听同一个rpc端点。
type Server struct {
	mu       sync.Mutex
	services map[string]*Service // 这个Server下面提供的全部服务
	count    int                 // 进入的rpc请求数量
}

// 创建一个服务器实例
func MakeServer() *Server {
	rs := &Server{
		services: map[string]*Service{},
	}
	return rs
}
func (rn *Network) AddServer(servername interface{}, rs *Server) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.servers[servername] = rs // 添加服务器
}

func (rn *Network) DeleteServer(servername interface{}) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.servers[servername] = nil // 删除服务器
}

func (rs *Server) dispatch(req reqMsg) replyMsg {
	rs.mu.Lock()

	rs.count += 1
	// 分离服务名字和方法
	dot := strings.LastIndex(req.svcMeth, ".")
	serviceName := req.svcMeth[:dot]
	methodName := req.svcMeth[dot+1:]

	// 获取服务方法需要的服务
	service, ok := rs.services[serviceName]

	rs.mu.Unlock()

	if ok {
		return service.dispatch(methodName, req)
	} else {
		var choices []string
		for k, _ := range rs.services {
			choices = append(choices, k)
		}
		// log输出全部的服务名字，还有不支持的服务（现在的请求）
		log.Fatalf("labrpc.Server.dispatch(): unknown service %v in %v.%v; expecting one of %v\n",
			serviceName, serviceName, methodName, choices)
		return replyMsg{false, nil}
	}
}

func (rs *Server) GetCount() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.count
}

//==============================Service=====================
// 带有方法的对象，可以通过rpc被调用，一个Server可以含有不止一个服务(Service)
type Service struct {
	name    string                    // 服务的名字
	rcvr    reflect.Value             //
	typ     reflect.Type              //
	methods map[string]reflect.Method // 方法的名字和方法的对应关系
}

func MakeService(rcvr interface{}) *Service {
	// 初始化一个服务实例
	svc := &Service{
		typ:     reflect.TypeOf(rcvr),
		rcvr:    reflect.ValueOf(rcvr),
		name:    reflect.Indirect(reflect.ValueOf(rcvr)).Type().Name(),
		methods: map[string]reflect.Method{},
	}
	for m := 0; m < svc.typ.NumMethod(); m++ {
		method := svc.typ.Method(m)
		mtype := method.Type
		mname := method.Name

		fmt.Printf("methodName:%v pkgPath:%v mTypeNumIn:%v IK1:%v IK2:%v  NumOut:%v\n",
			mname, method.PkgPath, mtype.NumIn(), mtype.In(1).Kind(), mtype.In(2).Kind(), mtype.NumOut())

		if method.PkgPath != "" || // capitalized?
			mtype.NumIn() != 3 ||
			mtype.In(2).Kind() != reflect.Ptr ||
			mtype.NumOut() != 0 {
			// the method is not suitable for a handler
			fmt.Printf("bad method: %v\n", mname)
		} else {
			// the method looks like a handler
			svc.methods[mname] = method
		}
	}
	return svc
}

func (rs *Server) AddService(svc *Service) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.services[svc.name] = svc
}

func (svc *Service) dispatch(methodName string, req reqMsg) replyMsg {
	//  根据方法名字，查找方法
	if method, ok := svc.methods[methodName]; ok {
		// prepare space into which to read the argument.
		// the Value's type will be a pointer to req.argsType.
		// 读取方法的参数
		paramTypes := reflect.New(req.argsType)
		// decode the argument.
		// 编码参数
		paramBytes := bytes.NewBuffer(req.args)
		paramDecoder := gob.NewDecoder(paramBytes)
		paramDecoder.Decode(paramTypes.Interface())

		// allocate space for the reply.
		// 分配响应信息的内存空间
		replyType := method.Type.In(2)
		replyType = replyType.Elem()
		replyv := reflect.New(replyType)

		// call the method.
		// 调用Call方法
		function := method.Func
		function.Call([]reflect.Value{svc.rcvr, paramTypes.Elem(), replyv})

		// encode the reply.
		// 解码响应信息
		rb := new(bytes.Buffer)
		re := gob.NewEncoder(rb)
		re.EncodeValue(replyv)

		return replyMsg{true, rb.Bytes()}
	} else {
		choices := []string{}
		for k, _ := range svc.methods {
			choices = append(choices, k)
		}
		log.Fatalf("labrpc.Service.dispatch(): unknown method %v in %v; expecting one of %v\n",
			methodName, req.svcMeth, choices)
		return replyMsg{false, nil}
	}
}

//============================services end====================

//==============================client========================
// 创建一个客户端,开启“线程”监听和分发
//端点
type ClientEnd struct {
	endName interface{}
	ch      chan reqMsg
}

func (rn *Network) MakEnd(endName interface{}) *ClientEnd {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if _, ok := rn.ends[endName]; ok {
		log.Fatalf("MakeEnd: %v already exists\n", endName)
	}

	e := &ClientEnd{
		endName: endName,
		ch:      rn.endCh, // 通过向e.ch发送消息就是往网络上面发送信息
	}
	rn.ends[endName] = e
	rn.enabled[endName] = false
	rn.connections[endName] = nil
	return e
}

func (e *ClientEnd) Call(svcMeth string, args interface{}, reply interface{}) bool {
	req := reqMsg{
		endName: e.endName,
		svcMeth: svcMeth,
		// 使用反射查找参数类型
		argsType: reflect.TypeOf(args),
		replyCh:  make(chan replyMsg),
	}
	// 序列化参数
	qb := new(bytes.Buffer)
	qe := gob.NewEncoder(qb)
	qe.Encode(args)

	req.args = qb.Bytes()
	// 发送请求
	e.ch <- req // 引用Network.endCh，会在Network中得到处理

	rep := <-req.replyCh // 获取回复,反序列化结果
	if rep.ok {
		rb := bytes.NewBuffer(rep.reply)
		rd := gob.NewDecoder(rb)
		if err := rd.Decode(reply); err != nil {
			log.Fatalf("ClientEnd.Call(): decode reply: %v\n", err)
		}
		return true
	} else {
		return false
	}
	return false
}