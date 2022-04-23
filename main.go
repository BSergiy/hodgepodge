package main

import (
	"fmt"
	"time"

	"github.com/BSergiy/hodgepodge/parallels/pools"
)

func main() {
	stop := make(chan int)
	defer close(stop)

	rl := pools.RateLimiter{
		WorkersNumber: 15,
		Rpm:           10,
		Command:       stop,
	}

	timer := func(command int) {
		stopTime := time.Tick(time.Second * 5)

		<-stopTime
		stop <- command
	}

	go timer(pools.StopAndJoin)

	if err := rl.Run(); err != nil {
		fmt.Println(err)
	}

	time.Sleep(3 * time.Second)

	go timer(pools.StopAndDetach)

	if err := rl.Run(); err != nil {
		fmt.Println(err)
	}
}
