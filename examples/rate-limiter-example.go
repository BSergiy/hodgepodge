package examples

import (
	"fmt"
	"time"

	"github.com/BSergiy/hodgepodge/parallels/pools"
)

func ExampleWithDetachAllWorkers() {
	example(pools.StopAndDetach)
}

func ExampleWithJoinAllWorkers() {
	example(pools.StopAndJoin)
}

func example(command int) {
	stop := make(chan int)
	defer close(stop)

	rl := pools.RateLimiter{
		WorkersNumber: 15,
		Rpm:           10,
		Command:       stop,
	}

	timer := func(command int) {
		stopTime := time.NewTicker(time.Second * 5)

		<-stopTime.C
		stop <- command

		stopTime.Stop()
	}

	go timer(command)

	if err := rl.Run(); err != nil {
		fmt.Println(err)
	}
}
