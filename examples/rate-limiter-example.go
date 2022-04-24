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

func example(command pools.Command) {
	commandsChannel := make(chan pools.Command)
	defer close(commandsChannel)

	rl, jobsChannel := pools.MakeRateLimiter(10, 100, commandsChannel)
	defer close(jobsChannel)

	timer := func(command pools.Command) {
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

		commandsChannel <- pools.ResetQueue

		timer.Reset(time.Second * 5)
		<-timer.C

		commandsChannel <- command

		timer.Stop()
	}

	go timer(command)

	if err := rl.Run(); err != nil {
		log.Println(err)
	}
}
