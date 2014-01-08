package rpcwrapper

import (
	"net/rpc"
)

// NewLocalClient creates an RPC client connected to the local RPC server srv.
func NewLocalClient(srv *rpc.Server) *rpc.Client {
	serverPipe, clientPipe := newBidiPipe()
	go srv.ServeConn(serverPipe)
	return rpc.NewClient(clientPipe)
}
