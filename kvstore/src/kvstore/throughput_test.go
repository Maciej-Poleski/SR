package kvstore_test

import (
	"fmt"
	"math/rand"
	"rpcwrapper"
	"sync"
	"testing"
	"time"
)

func laggyNetwork(c rpcwrapper.Client) rpcwrapper.Client {
	return &rpcwrapper.InterceptedClient{
		Client: c,
		Interceptor: func() rpcwrapper.InterceptResult {
			time.Sleep(100*time.Millisecond + time.Duration(rand.Intn(30))*time.Millisecond)
			return rpcwrapper.Success
		},
	}
}

func laggyLossyNetwork(c rpcwrapper.Client) rpcwrapper.Client {
	return &rpcwrapper.InterceptedClient{
		Client: c,
		Interceptor: func() rpcwrapper.InterceptResult {
			time.Sleep(100*time.Millisecond + time.Duration(rand.Intn(30))*time.Millisecond)
			if rand.Intn(4) == 0 {
				return rpcwrapper.FailedRequest
			}
			if rand.Intn(4) == 0 {
				return rpcwrapper.FailedResponse
			}
			return rpcwrapper.Success
		},
	}
}

func TestThroughput(t *testing.T) {
	testThroughput(t, laggyNetwork)
}

func TestThroughputLossy(t *testing.T) {
	testThroughput(t, laggyLossyNetwork)
}

func testThroughput(t *testing.T, network func(rpcwrapper.Client) rpcwrapper.Client) {
	master, slaves, quit := createService(3, network)
	defer close(quit)

	const N = 600

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			set(t, master, "k", fmt.Sprint(i))
			wg.Done()
		}(i)
	}
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
			set(t, master, "k", fmt.Sprint(i))
			wg.Done()
		}(i)
	}
	wg.Wait()

	v := get(t, master, "k")
	if v == "" {
		t.Errorf("Value empty after many Set calls")
	}
	for _, s := range slaves {
		getExpect(t, s, "k", v)
	}
}
