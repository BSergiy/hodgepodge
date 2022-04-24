package pools

import (
	"log"
	"sync"
	"time"
)

type worker struct {
	JobsChanel chan func()
}

func (worker *worker) run(awaiter *sync.WaitGroup) {
	awaiter.Add(1)
	defer awaiter.Done()

	for job := range worker.JobsChanel {
		log.Printf("Job started...\n")

		started := time.Now()
		job()
		log.Printf("Job finished (taked %v)\n", time.Since(started))
	}

	log.Println("--- worker was destructed")
}
