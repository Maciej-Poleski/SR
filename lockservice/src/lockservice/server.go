// +build !sol

package lockservice

import (
	"sync"
)

type ClientState struct {
	RequestId    int
	lockResponse bool
}

type LockService struct {
	mu      sync.Mutex
	locks   map[string]bool
	clients map[int]ClientState
}

func NewLockService() interface{} {
	return &LockService{
		locks:   make(map[string]bool),
		clients: make(map[int]ClientState),
	}
}

func (ls *LockService) Lock(request *LockRequest, response *LockResponse) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	cs := ls.clients[request.ClientId]
	if cs.RequestId < request.RequestId {
		locked := ls.locks[request.LockName]
		if locked {
			response.Ok = false
		} else {
			ls.locks[request.LockName] = true
			response.Ok = true
		}
		cs.RequestId = request.RequestId
		cs.lockResponse = response.Ok
		ls.clients[request.ClientId] = cs
	} else {
		response.Ok = cs.lockResponse
	}
	return nil
}

func (ls *LockService) Unlock(request *UnlockRequest, response *UnlockResponse) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	cs := ls.clients[request.ClientId]
	if cs.RequestId < request.RequestId {
		cs.RequestId = request.RequestId
		ls.clients[request.ClientId] = cs
		delete(ls.locks, request.LockName)
	}
	return nil
}
