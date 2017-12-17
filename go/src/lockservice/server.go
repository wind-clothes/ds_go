package lockservice

import (
	"io"
	"net"
	"sync"
)

type LockServer struct {
	mu       sync.Mutex
	listener net.Listener
	dead     bool              // for test_test.go
	dying    bool              // for test_test.go

	isPrimary bool            // am I the primary?
	backup    string          // backup's port
	locks     map[string]bool // for each lock name, is it locked?
}

type DeafConn struct {
	reader io.ReadWriteCloser
}

func (ls *LockServer) StartServer() {

}


func (ls *LockServer) Lock(lockName string) bool  {
	return  false
}

func (ls *LockServer) UnLock(lockName string) bool {
	return  false
}

func (ls *LockServer) tryAcquire(lockName string) bool  {
	return false
}

func (ls *LockServer) tryRelease()  {

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
