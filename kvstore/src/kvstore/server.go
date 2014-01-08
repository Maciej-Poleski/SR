// +build !sol

package kvstore

import "rpcwrapper"
import "sync"

type Item struct {
	Key   string
	Value string        // Dane do zapisu
	Ch    chan struct{} // Zostanie zamknięty gdy to żądanie zostanie wykonane
	deg   int           // Pozostała ilość zablokowanych na kanale
	lock  sync.Mutex    // Blokada całego obiektu
}

type Store struct {
	lock      sync.Mutex          // Blokada całego obiektu
	log       map[string]string   // Dziennik
	version   int                 // Wersja dziennika (Master trzyma przecięcie po wszystkich slave)
	slaves    []rpcwrapper.Client // tylko w Masterze
	quit      chan struct{}       // Kanał kończący
	queued    map[int]*Item       // Lista żądań zakolejkowanych (czekających na wykonanie żądania o niższym id)
	nextReqId int                 // Identyfikator dla kolejnego żądania zapisu do slave
}

func NewMaster(slaves []rpcwrapper.Client, quit chan struct{}) interface{} {
	return &Store{
		log:       make(map[string]string),
		version:   0,
		slaves:    slaves[:],
		quit:      quit,
		queued:    make(map[int]*Item),
		nextReqId: 1,
	}
}

func NewSlave(quit chan struct{}) interface{} {
	return &Store{
		log:     make(map[string]string),
		version: 0,
		quit:    quit,
		queued:  make(map[int]*Item),
	}
}

func (s *Store) Get(req *GetRequest, resp *GetResponse) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.log[req.Key]
	if ok {
		resp.Value = &v
	} else {
		resp.Value = nil
	}
	return nil
}

type SlaveSetRequest struct {
	Key   string
	Value string
	Id    int
}

type SlaveSetResponse struct {
}

// Ustawia Slave (lub Mastera) (czekając na wprowadzenie do dziennika wcześniejszych wersji)
func (s *Store) SetSlave(req *SlaveSetRequest, resp *SlaveSetResponse) error {
	s.lock.Lock()
	if req.Id == s.version+1 {
		// Wykonaj
		s.version = req.Id
		s.log[req.Key] = req.Value
		for v, ok := s.queued[s.version+1]; ok; v, ok = s.queued[s.version+1] {
			// Sprawdź kolejke
			s.version = s.version + 1
			s.log[v.Key] = v.Value
			close(v.Ch)
		}
		s.lock.Unlock()
		return nil
	} else if req.Id > s.version+1 {
		if _, ok := s.queued[req.Id]; ok == false { // Być może już zakolejkowano
			s.queued[req.Id] = &Item{
				Key:   req.Key,
				Value: req.Value,
				Ch:    make(chan struct{}),
				deg:   0,
			}
		}
		s.queued[req.Id].lock.Lock()
		s.queued[req.Id].deg++
		s.queued[req.Id].lock.Unlock()
		s.lock.Unlock()
		<-s.queued[req.Id].Ch
		s.lock.Lock()
		s.queued[req.Id].lock.Lock()
		s.queued[req.Id].deg--
		if s.queued[req.Id].deg == 0 {
			//delete(s.queued, req.Id) // Nie zdejmuje blokady - obiekt i tak został skasowany
		} else {
			s.queued[req.Id].lock.Unlock()
		}
		s.lock.Unlock()
		return nil
	} else { // Żądanie zostało już wykonane
		s.lock.Unlock()
		return nil
	}
	panic("unreachable")
}

func (s *Store) Set(req *SetRequest, resp *SetResponse) error {
	storedChan := make(chan int)
	s.lock.Lock()
	reqId := s.nextReqId
	s.nextReqId += 1
	s.lock.Unlock()
	for i, v := range s.slaves {
		go func(i int, v rpcwrapper.Client) {
			for {
				select {
				default:
					reql := SlaveSetRequest{
						Key:   req.Key,
						Value: req.Value,
						Id:    reqId,
					}
					respl := SlaveSetResponse{}
					if err := v.Call("Store.SetSlave", &reql, &respl); err == nil {
						storedChan <- i
						return
					}
				case <-s.quit:
					return
				}
			}
			panic("unreachable")
		}(i, v)
	}
	for i := 0; i < len(s.slaves); i++ {
		<-storedChan
	}
	close(storedChan)
	reql := SlaveSetRequest{
		Key:   req.Key,
		Value: req.Value,
		Id:    reqId,
	}
	respl := SlaveSetResponse{}
	s.SetSlave(&reql, &respl)
	return nil
}
