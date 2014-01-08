package kvstore

type GetRequest struct {
	Key string
}

type GetResponse struct {
	Value *string
}

type SetRequest struct {
	Key   string
	Value string
}

type SetResponse struct {
}
