package main

import (
    "fmt"
    "log"
    "net"
    "net/rpc"
    "sync"
)

//
// Common between client and server
//

const (
    OK       = "OK"
    ErrNoKey = "ErrNoKey"
)

type Err string

type PutArgs struct {
    Key   string
    Value string
}

type PutReply struct {
    Err Err
}

type GetArgs struct {
    Key string
}

type GetReply struct {
    Err   Err
    Value string
}

// Server

// 存储的数据结构体
type KV struct {
    mu       sync.Mutex
    keyvalue map[string]string
}

// 通过key值从KV中获取值
func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
    kv.mu.Lock()
    defer kv.mu.Unlock()

    reply.Err = "OK"
    // map方法的返回值为两个一个是值，一个是是否存在
    val, ok := kv.keyvalue[args.Key]
    if ok {
        reply.Err = OK
        reply.Value = val
    } else {
        reply.Err = ErrNoKey
        reply.Value = ""
    }
    return nil
}

// 往KV中存储数据
func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
    kv.mu.Lock()
    defer kv.mu.Unlock()

    kv.keyvalue[args.Key] = args.Value
    reply.Err = OK
    return nil
}

// server 
func server() {
    kv := new(KV)
    kv.keyvalue = map[string]string{}
    rpcs := rpc.NewServer
    rpcs.Register(kv)

    l, e := net.listen("tcp",":1234")
    if e != nil {
        log.Fatal("listen error:", e)
    }
    // 启动协程来进行通信
    go func() {
        for {
            conn, err := l.Accept()
            if err == nil {
                go rpcs.ServeConn(conn)
            } else {
                break
            }
        }
        l.Close()
        fmt.Printf("Server done\n")
    }
}

// Client

func Dial() *rpc.Client {
    client, err := rpc.Dial("tcp", ":1234")
    if err != nil {
        log.Fatal("dialing:", err)
    }
    return client
}

func Get(key string) string {
    client := Dial()
    args := &GetArgs{"subject"}
    reply := GetReply{"", ""}
    err := client.Call("KV.Get", args, &reply)
    if err != nil {
        log.Fatal("error:", err)
    }
    client.Close()
    return reply.Value
}

func Put(key string, val string) {
    client := Dial()
    args := &PutArgs{"subject", "6.824"}
    reply := PutReply{""}
    err := client.Call("KV.Put", args, &reply)
    if err != nil {
        log.Fatal("error:", err)
    }
    client.Close()
}

// main
func main() {
    server()

    Put("subject", "6.824")
    fmt.Printf("Put subjet done\n")
    fmt.Printf("Get %s\n", Get("subject"))
}