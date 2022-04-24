package examples

import (
	"log"
	"math/rand"
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
		WorkersNumber: 10,
		JobsPerMin:    100,
		Command:       stop,
	}

	jobsChannel, err := rl.MakeJobQueueChannel()

	if err != nil {
		log.Println(err)
		return
	}

	timer := func(command int) {
		timer := time.NewTicker(time.Second * 2)

		<-timer.C

		jobs := make([]func(), 0)

		for i := 0; i < 10; i++ {
			j := i
			jobs = append(jobs, func() {
				minDelay := 0
				maxDelay := 1

				time.Sleep(time.Duration(rand.Intn(maxDelay-minDelay)+minDelay) * time.Second)
				log.Println("Job #", j)
			})
		}

		jobsChannel <- jobs

		timer.Reset(time.Second * 5)

		<-timer.C
		stop <- command

		timer.Stop()
	}

	go timer(command)

	if err := rl.Run(); err != nil {
		log.Println(err)
	}
}
