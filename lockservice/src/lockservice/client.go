// +build !sol

package lockservice

import (
	"math/rand"
	"rpcwrapper"
)

type Client struct {
	srv       rpcwrapper.Client
	ClientId  int
	RequestId int
}

func NewClient(rpcClient rpcwrapper.Client) *Client {
	return &Client{rpcClient, rand.Int(), 0}
}

func (c *Client) Lock(lockName string) bool {
	c.RequestId++
	for {
		req := LockRequest{
			LockName:  lockName,
			ClientId:  c.ClientId,
			RequestId: c.RequestId,
		}
		resp := LockResponse{}
		if err := c.srv.Call("LockService.Lock", &req, &resp); err == nil {
			return resp.Ok
		}
	}
}

func (c *Client) Unlock(lockName string) {
	c.RequestId++
	for {
		req := UnlockRequest{
			LockName:  lockName,
			ClientId:  c.ClientId,
			RequestId: c.RequestId,
		}
		resp := UnlockResponse{}
		if err := c.srv.Call("LockService.Unlock", &req, &resp); err == nil {
			return
		}
	}
}

func (c *Client) Close() {
	c.srv.Close()
}
