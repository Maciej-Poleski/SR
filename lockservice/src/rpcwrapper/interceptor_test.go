package rpcwrapper

import (
	"net/rpc"
	"sync"
	"testing"
	"time"
)

type TestService struct {
	A chan bool
}

func (ts *TestService) DoA(request *struct{}, response *bool) error {
	*response = <-ts.A
	return nil
}

func clientServerPair(interceptor func() InterceptResult) (*TestService, Client) {
	srv := rpc.NewServer()
	ts := &TestService{}
	srv.Register(ts)
	client := &InterceptedClient{
		Client: NewLocalClient(srv),
		Interceptor: interceptor,
	}
	return ts, client
}

func TestPassthrough(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ts, client := clientServerPair(func() InterceptResult { return Success })
	defer client.Close()
	ts.A = make(chan bool)

	doneWait := false
	wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		doneWait = true
		ts.A <- true
		wg.Done()
	}()
	var resp bool
	if err := client.Call("TestService.DoA", &struct{}{}, &resp); err != nil {
		t.Fatalf("client.Call: %s", err.Error())
	}
	if resp != true {
		t.Fatalf("TestService.DoA: expected true, got false")
	}
	if !doneWait {
		t.Fatalf("got response before the service responded")
	}
}

func TestFailRequest(t *testing.T) {
	_, client := clientServerPair(func() InterceptResult { return FailedRequest })
	defer client.Close()

	var resp bool
	if err := client.Call("TestService.DoA", &struct{}{}, &resp); err == nil {
		t.Fatalf("no error from client.Call with FailRequest interceptor")
	}
}

func TestFailResponse(t *testing.T) {
	ts, client := clientServerPair(func() InterceptResult { return FailedResponse })
//	defer client.Close()
	ts.A = make(chan bool, 1)
	ts.A <- true

	resp := false
	if err := client.Call("TestService.DoA", &struct{}{}, &resp); err == nil {
		t.Fatalf("no error from client.Call with FailedResponse interceptor")
	}
	if resp != false {
		t.Fatalf("response leaked through when interceptor specified FailedResponse")
	}
	// The request can reach the server after we've gotten a failure response, but should reach it sometime.
	time.Sleep(100 * time.Millisecond)
	if len(ts.A) != 0 {
		t.Fatalf("no request received on the server when interceptor specified FailedResponse intercept result %v", len(ts.A))
	}
}

