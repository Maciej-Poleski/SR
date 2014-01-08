package rpcwrapper

import (
	"errors"
	"log"
	"net/rpc"
	"reflect"
)

type InterceptResult int

const (
	Success InterceptResult = iota
	FailedRequest
	FailedResponse
)

// An InterceptedClient is a wrapper around client that passes each call to the wrapped Client
// through the Interceptor function. The function is free to delay calls, and then allow them
// to complete successfully, cause them to fail immediately, or cause them to be actually executed
// while seemingly failing.
type InterceptedClient struct {
	Client
	Interceptor func() InterceptResult
}

var ErrNoResponse = errors.New("rpcwrapper: simulated lack of response")

func (ic *InterceptedClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	switch ic.Interceptor() {
	case Success:
		return ic.Client.Call(serviceMethod, args, reply)
	case FailedRequest:
		return ErrNoResponse
	case FailedResponse:
		replyType := reflect.TypeOf(reply)
		if replyType.Kind() != reflect.Ptr {
			log.Print("rpcwrapper: Call got passed a nonpointer as the reply")
		}
		// If we missed the response, we could just as well return before the server has finished responding.
		go func() {
			dummy := reflect.New(replyType.Elem()).Interface()
			// This call might actually happen after the client is closed. It's perfectly fine and thus, we ignore all errors here.
			ic.Client.Call(serviceMethod, args, dummy)
		}()
		return ErrNoResponse
	default:
		panic("invalid interceptor result")
	}
	panic("unreachable")
}

func (ic *InterceptedClient) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	if done == nil {
		done = make(chan *rpc.Call, 1)
	}
	if cap(done) == 0 {
		log.Panic("rpcwrapper: done channel is unbuffered")
	}
	call := &rpc.Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Error:         nil,
		Done:          done,
	}
	go func() {
		call.Error = ic.Call(serviceMethod, args, reply)
		call.Done <- call
	}()
	return call
}
