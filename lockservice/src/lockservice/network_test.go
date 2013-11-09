package lockservice_test

import (
	"fmt"
	"lockservice"
	"math/rand"
	"reflect"
	"rpcwrapper"
	"sync"
	"sync/atomic"
	"testing"
	"testing/quick"
	"time"
)

func checkLock(t *testing.T, c *lockservice.Client, lockName string, expectedResult bool) {
	if result := c.Lock(lockName); result != expectedResult {
		t.Fatalf("Lock(%s): expected %v, got %v", lockName, expectedResult, result)
	}
}

// The following tests leak goroutines (the clients never close their connections, so the servers are left hanging
// trying to read from the already gone clients).

type interceptSequence []rpcwrapper.InterceptResult

func (_ interceptSequence) Generate(rand *rand.Rand, size int) reflect.Value {
	for rand.Intn(2) == 0 {
		size--
	}
	is := make(interceptSequence, size)
	for i := 0; i < size; i++ {
		if rand.Intn(2) == 0 {
			is[i] = rpcwrapper.FailedRequest
		} else {
			is[i] = rpcwrapper.FailedResponse
		}
	}
	return reflect.ValueOf(is)
}

func TestCanAlwaysLock(t *testing.T) {
	var mode chan rpcwrapper.InterceptResult
	ls := newService()
	currentMode := func() rpcwrapper.InterceptResult {
		select {
		case m := <-mode:
			return m
		default:
			return rpcwrapper.Success
		}
		panic("unreachable")
	}
	cs := []*lockservice.Client{
		newClient(ls, currentMode),
		newClient(ls, currentMode),
	}
	defer cs[0].Close()
	defer cs[1].Close()
	modes := []interceptSequence{
		{},
		{rpcwrapper.FailedRequest},
		{rpcwrapper.FailedResponse},
		{rpcwrapper.FailedRequest, rpcwrapper.FailedResponse},
		{rpcwrapper.FailedResponse, rpcwrapper.FailedResponse, rpcwrapper.FailedRequest},
	}
	testSequence := func(is interceptSequence) bool {
		i := rand.Intn(2)
		mode = make(chan rpcwrapper.InterceptResult, len(is))
		for _, m := range is {
			mode <- m
		}
		if result := cs[i%2].Lock("a"); result != true {
			t.Errorf("Lock(a) failed with mode sequence %+v", is)
			return false
		}
		mode = make(chan rpcwrapper.InterceptResult, len(is))
		for _, m := range is {
			mode <- m
		}
		cs[(i+1)%2].Unlock("a")
		mode = nil
		if cs[0].Lock("a") != true {
			t.Errorf("Lock(a) failed after Unlock(a) with mode sequence %+v", is)
			return false
		}
		cs[0].Unlock("a")
		return true
	}
	for _, ms := range modes {
		if !testSequence(ms) {
			return
		}
	}
	if err := quick.Check(testSequence, nil); err != nil {
		if _, checkErr := err.(*quick.CheckError); !checkErr {
			panic(err)
		}
	}
}

func unreliableMode() rpcwrapper.InterceptResult {
	r := rand.Float32()
	if r < 0.05 {
		return rpcwrapper.FailedRequest
	} else if r < 0.1 {
		return rpcwrapper.FailedResponse
	} else {
		return rpcwrapper.Success
	}
	panic("unreachable")
}

func TestUnreliableSameLock(t *testing.T) {
	const numClients = 10
	ls := newService()
	var wg sync.WaitGroup
	var quitFlag int32
	var lockCount, unlockCount int32
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func() {
			c := newClient(ls, unreliableMode)
			defer c.Close()
			for atomic.LoadInt32(&quitFlag) == 0 {
				ok := c.Lock("a")
				if ok {
					atomic.AddInt32(&lockCount, 1)
				}
				if atomic.LoadInt32(&quitFlag) != 0 {
					break
				}
				if ok {
					c.Unlock("a")
					atomic.AddInt32(&unlockCount, 1)
				}
			}
			wg.Done()
		}()
	}
	time.Sleep(time.Second)
	atomic.StoreInt32(&quitFlag, 1)
	wg.Wait()
	c := newClient(ls, unreliableMode)
	defer c.Close()
	if c.Lock("a") {
		lockCount++
	}
	if lockCount != unlockCount+1 {
		t.Fatalf("wrong count of locks and unlocks: %v locks and %v unlocks", lockCount, unlockCount)
	}
}

func TestUnreliableManyLocks(t *testing.T) {
	const numClients = 10
	const numLocks = 10
	ls := newService()
	lockName := func(clientNum int, lockNum int) string {
		return fmt.Sprintf("%d-%d", clientNum, lockNum)
	}
	var wg sync.WaitGroup
	var quitFlag int32
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientNum int) {
			c := newClient(ls, unreliableMode)
			defer c.Close()
			for atomic.LoadInt32(&quitFlag) == 0 {
				for j := 0; j < numLocks; j++ {
					checkLock(t, c, lockName(clientNum, j), true)
					c.Unlock(lockName(clientNum, j))
				}
			}
			wg.Done()
		}(i)
	}
	time.Sleep(time.Second)
	atomic.StoreInt32(&quitFlag, 1)
	wg.Wait()
	c := newClient(ls, unreliableMode)
	defer c.Close()
	for i := 0; i < numClients; i++ {
		for j := 0; j < numLocks; j++ {
			checkLock(t, c, lockName(i, j), true)
		}
	}
}
