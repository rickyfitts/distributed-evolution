package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/rickyfitts/distributed-evolution/api"
)

func (m *Master) Echo(args *api.EchoArgs, reply *api.EchoReply) error {
	fmt.Printf("request received: { Message: %v }\n", args.Message)
	reply.Message = args.Message
	return nil
}

func (m *Master) rpcServer() {
	rpc.Register(m)
	rpc.HandleHTTP()

	port := os.Getenv("RPC_PORT")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listener error: ", err)
	}

	fmt.Printf("rpc listening on port %v\n", port)

	http.Serve(listener, nil)
}
