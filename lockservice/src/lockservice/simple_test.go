package lockservice_test

import (
	"lockservice"
	"rpcwrapper"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSmoke(t *testing.T) {
	ls := newService()
	c := newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success })
	defer c.Close()
	for _, op := range []struct {
		fn     string
		lock   string
		result bool
	}{
		{"Lock", "a", true},
		{"Lock", "a", false},
		{"Lock", "b", true},
		{"Unlock", "a", true},
		{"Lock", "a", true},
	} {
		if op.fn == "Lock" {
			if result := c.Lock(op.lock); result != op.result {
				t.Fatalf("Lock(%s): expected %v, got %v", op.lock, op.result, result)
			}
		} else if op.fn == "Unlock" {
			c.Unlock(op.lock)
		} else {
			t.Fatalf("unknown lock operation: %s", op.fn)
		}
	}
}

func TestMultipleClients(t *testing.T) {
	ls := newService()
	c := []*lockservice.Client{
		newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success }),
		newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success }),
	}
	defer c[0].Close()
	defer c[1].Close()
	if c[0].Lock("a") != true {
		t.Fatalf("Lock(a): expected true, got false")
	}
	if c[1].Lock("a") != false {
		t.Fatalf("Lock(a): expected false, got true")
	}
	c[1].Unlock("a")
	if c[0].Lock("a") != true {
		t.Fatalf("Lock(a): expected true, got false")
	}
}

func TestConcurrentClients(t *testing.T) {
	ls := newService()
	cs := []*lockservice.Client{
		newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success }),
		newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success }),
		newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success }),
		newClient(ls, func() rpcwrapper.InterceptResult { return rpcwrapper.Success }),
	}
	for _, c := range cs {
		defer c.Close()
	}
	var wg sync.WaitGroup
	var successes int32
	for _, c := range cs {
		wg.Add(1)
		go func(client *lockservice.Client) {
			if client.Lock("a") {
				atomic.AddInt32(&successes, 1)
			}
			wg.Done()
		}(c)
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("%d processes succeeded in simultaneously locking a single lock", successes)
	}
}
