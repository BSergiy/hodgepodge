package pools

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	Ch  chan int
	Job func()
}

func (worker *Worker) Run(awaiter *sync.WaitGroup) {
	awaiter.Add(1)
	defer awaiter.Done()

	for jobNumber := range worker.Ch {
		fmt.Printf("Job #%d started...\n", jobNumber)

		started := time.Now()
		worker.Job()
		fmt.Printf("Job #%d finished (taked %v)\n", jobNumber, time.Since(started))
	}

	fmt.Println("--- worker was destructed")
}
