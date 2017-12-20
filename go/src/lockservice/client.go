package lockservice

import (
	"fmt"
	"net/rpc"
)

type LockClient struct {
	servers [2]string // primary port, backup port
}

func MakeLockClient(primary string, backup string) (lc *LockClient) {
	lc = &LockClient{
		servers: [2]string{primary, backup},
	}
	return lc
}

func (lc *LockClient) Lock(lockname string) bool {
	lockParam := &LockArgs{
		Lockname: lockname,
	}

	lockReply := &LockReply{}

	ok := call(lc.servers[0], "LockServer.Lock", lockParam, lockReply)
	if ok {
		return lockReply.OK
	} else {
		return false
	}
}

func (lc *LockClient) UnLock(lockname string) bool {
	lockParam := &LockArgs{
		Lockname: lockname,
	}

	lockReply := &LockReply{}

	ok := call(lc.servers[0], "LockServer.UnLock", lockParam, lockReply)
	if ok {
		return lockReply.OK
	} else {
		return false
	}
}
func call(server string, rpcname string,
	args interface{}, reply interface{}) bool {
	client, err := rpc.Dial("unix", server)

	if err != nil {
		return false
	}
	defer client.Close()

	error := client.Call(rpcname, args, reply)
	if error != nil {
		fmt.Print(error)
		return false
	}
	return true
}
