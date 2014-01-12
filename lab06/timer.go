package main

import "fmt"
import "sync"
import "time"
import "math/rand"

type Timer struct {
	ch   chan int
	time int
	lock sync.Mutex
}

var n = 10
var timers = make([]Timer, n)

func timerReceiver(i int) {
	for {
		x := <-timers[i].ch
		timers[i].lock.Lock()
		if timers[i].time < x {
			timers[i].time = x
			fmt.Printf("Time on %d is %d\n", i, x)
		}
		timers[i].lock.Unlock()
	}
}

func timerSender(i int) {
	timers[i].lock.Lock()
	timers[i].time++
	timers[i].lock.Unlock()
	for j := 0; j < n; j++ {
		if j == i {
			continue
		}
		timers[j].ch <- timers[i].time
	}
}

func main() {

	for i := 0; i < n; i++ {
		timers[i] = Timer{ch: make(chan int), time: 0}
	}
	for i := 0; i < n; i++ {
		go timerReceiver(i)
	}
	for {
		for i := 0; i < n; i++ {
			time.Sleep(time.Duration(100+rand.Intn(300)) * time.Millisecond)
			timerSender(i)
		}
	}
}
