// The package rpcwrapper contains interfaces that icap or mimic net/rpc functionality,
// while allowing the calls to be intercepted.
package rpcwrapper

import (
	"net/rpc"
)

// A Client is an interface exactly equal to the public interface of rpc.Client.
type Client interface {
	Call(serviceMethod string, args interface{}, reply interface{}) error
	Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call
	Close() error
}
