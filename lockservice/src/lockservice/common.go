// +build !sol

package lockservice

type LockRequest struct {
	LockName  string
	ClientId  int
	RequestId int
}

type LockResponse struct {
	Ok bool
}

type UnlockRequest struct {
	LockName  string
	ClientId  int
	RequestId int
}

type UnlockResponse struct {
}
