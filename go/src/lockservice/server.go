package lockservice

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type LockServer struct {
	mu       sync.Mutex
	listener net.Listener
	dead     bool // for test_test.go
	dying    bool // for test_test.go

	isPrimary bool            // am I the primary?
	backup    string          // backup's port
	locks     map[string]bool // for each lock name, is it locked?
}

type DeafConn struct {
	reader io.ReadWriteCloser
}

func StartServer(primary string, backup string, isPrimary bool) *LockServer {
	ls := &LockServer{
		isPrimary: isPrimary,
		backup:    backup,
		locks:     map[string]bool{},
	}
	me := ""
	if isPrimary {
		me = primary
	} else {
		me = backup
	}

	// tell net/rpc about our RPC server and handlers.
	rpcs := rpc.NewServer()
	rpcs.Register(ls)

	// prepare to receive connections from clients.
	// change "unix" to "tcp" to use over a network.
	os.Remove(me) // only needed for "unix"
	l, e := net.Listen("unix", me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	ls.listener = l

	// please don't change any of the following code,
	// or do anything to subvert it.

	// create a thread to accept RPC connections from clients.
	go func() {
		for ls.dead == false {
			conn, err := ls.listener.Accept()
			if err == nil && ls.dead == false {
				if ls.dying {
					go func() {
						time.Sleep(2 * time.Second)
						conn.Close()
					}()
					ls.listener.Close()

					deafConn := DeafConn{
						reader: conn,
					}
					rpc.ServeConn(deafConn)

					ls.dead = true
				} else {
						go rpcs.ServeConn(conn)
				}
			} else if err == nil {
				conn.Close()
			}
			if err != nil && ls.dead == false {
				fmt.Printf("LockServer(%v) accept: %v\n", me, err.Error())
				ls.kill()
			}
		}
	}()

	return ls
}

func (ls *LockServer) kill() {
	ls.dead = true
	ls.listener.Close()
}

func (ls *LockServer) Lock(lockParam *LockArgs, reply *LockReply) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	locked, _ := ls.locks[lockParam.Lockname]

	if locked {
		reply.OK = false
	} else {
		reply.OK = true
		ls.locks[lockParam.Lockname] = true
	}
	return nil
}

func (ls *LockServer) UnLock(lockParam *LockArgs, reply *LockReply) error {
	ls.mu.Lock()
	defer  ls.mu.Unlock()

	locked, _ := ls.locks[lockParam.Lockname]

	if locked {
		reply.OK = true
		ls.locks[lockParam.Lockname] = false
	}else {
		reply.OK = false
	}
	return nil
}

func (ls *LockServer) tryAcquire(lockParam *LockArgs) bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	locked, _ := ls.locks[lockParam.Lockname]
	if locked {
		return false
	}
	return true
}

func (ls *LockServer) tryRelease(lockParam *LockArgs) bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	locked, _ := ls.locks[lockParam.Lockname]

	if locked {
		return true
	}
	return false
}

func (dc DeafConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (dc DeafConn) Close() error {
	return dc.reader.Close()
}

func (dc DeafConn) Read(p []byte) (n int, err error) {
	return dc.reader.Read(p)
}
