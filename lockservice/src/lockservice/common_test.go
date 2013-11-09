package lockservice_test

import (
	"lockservice"
	"net/rpc"
	"rpcwrapper"
)

func newService() *rpc.Server {
	server := rpc.NewServer()
	ls := lockservice.NewLockService()
	server.Register(ls)
	return server
}

func newClient(srv *rpc.Server, interceptor func() rpcwrapper.InterceptResult) *lockservice.Client {
	return lockservice.NewClient(&rpcwrapper.InterceptedClient{
		Client:      rpcwrapper.NewLocalClient(srv),
		Interceptor: interceptor,
	})
}
