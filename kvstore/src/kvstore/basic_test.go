package kvstore_test

import (
	"kvstore"
	"net/rpc"
	"rpcwrapper"
	"testing"
)

func get(t *testing.T, s rpcwrapper.Client, key string) string {
	req := kvstore.GetRequest{key}
	resp := kvstore.GetResponse{}
	err := s.Call("Store.Get", &req, &resp)
	if err != nil {
		t.Errorf("get(%s) returned an error: %v", key, err)
	}
	if resp.Value == nil {
		return ""
	}
	return *resp.Value
}

func getExpect(t *testing.T, s rpcwrapper.Client, key string, expectedValue string) {
	value := get(t, s, key)
	if expectedValue != value {
		t.Errorf("get(%s) = '%s', not '%s'", key, value, expectedValue)
	}
}

func set(t *testing.T, s rpcwrapper.Client, key string, value string) {
	req := kvstore.SetRequest{key, value}
	resp := kvstore.SetResponse{}
	err := s.Call("Store.Set", &req, &resp)
	if err != nil {
		t.Errorf("get(%s) returned an error: %v", key, err)
	}
}

func idealNetwork(c rpcwrapper.Client) rpcwrapper.Client {
	return c
}

func createService(slaveCount int, networkSim func(rpcwrapper.Client) rpcwrapper.Client) (master rpcwrapper.Client, slaves []rpcwrapper.Client, quit chan struct{}) {
	quit = make(chan struct{})
	slaves = make([]rpcwrapper.Client, slaveCount)
	for i := range slaves {
		srv := rpc.NewServer()
		srv.RegisterName("Store", kvstore.NewSlave(quit))
		slaves[i] = rpcwrapper.NewLocalClient(srv)
		go func(c rpcwrapper.Client) {
			<-quit
			c.Close()
		}(slaves[i])
	}
	mastersSlaves := make([]rpcwrapper.Client, slaveCount)
	for i, s := range slaves {
		mastersSlaves[i] = networkSim(s)
	}
	srv := rpc.NewServer()
	srv.RegisterName("Store", kvstore.NewMaster(mastersSlaves, quit))
	master = rpcwrapper.NewLocalClient(srv)
	go func(c rpcwrapper.Client) {
		<-quit
		c.Close()
	}(master)
	return
}

func TestSmokeSingle(t *testing.T) {
	master, _, quit := createService(0, idealNetwork)
	defer close(quit)

	getExpect(t, master, "k1", "")
	set(t, master, "k1", "v1")
	getExpect(t, master, "k1", "v1")
	set(t, master, "k2", "v2")
	getExpect(t, master, "k1", "v1")
	getExpect(t, master, "k2", "v2")
	getExpect(t, master, "k3", "")
}

func TestSmokeBackups(t *testing.T) {
	master, slaves, quit := createService(3, idealNetwork)
	defer close(quit)

	getExpect(t, master, "k1", "")
	set(t, master, "k1", "v1")
	getExpect(t, master, "k1", "v1")
	set(t, master, "k2", "v2")
	getExpect(t, master, "k1", "v1")
	getExpect(t, master, "k2", "v2")
	getExpect(t, master, "k3", "")

	for _, s := range slaves {
		getExpect(t, s, "k1", "v1")
		getExpect(t, s, "k2", "v2")
		getExpect(t, s, "k3", "")
	}
}
