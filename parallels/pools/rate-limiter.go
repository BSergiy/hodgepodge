package pools

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"
)

/// TODO: make tests
/// TODO: implement internal timer for execute queue
/// TODO: improve worker and job identification mechanism

type RateLimiter struct {
	WorkersNumber int
	JobsPerMin    int
	Command       chan int

	awaiter             sync.WaitGroup
	internalJobsChannel chan func()
	safeClose           func()
	bucket              int
	ticker              *time.Ticker

	jobQueueChannel chan []func()
	jobQueue        []func()
	nextJobIndex    int
}

func (limiter *RateLimiter) MakeJobQueueChannel() (chan []func(), error) {
	if limiter.jobQueueChannel != nil {
		return nil, errors.New("job queue already created")
	}

	limiter.jobQueueChannel = make(chan []func())

	return limiter.jobQueueChannel, nil
}

func (limiter *RateLimiter) Run() error {
	if err := limiter.checkOnErrors(); err != nil {
		return err
	}

	limiter.prepareState()

	defer limiter.ticker.Stop()
	defer limiter.safeClose()

	for {
		if err, canceled := limiter.checkChannels(); canceled {
			return err
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func (limiter *RateLimiter) checkOnErrors() error {
	if limiter.JobsPerMin <= 0 {
		return errors.New("'JobsPerMin' cannot be less than 1")
	}
	if limiter.jobQueueChannel == nil {
		return errors.New("you must execute 'MakeJobQueueChannel' before 'Run'")
	}

	return nil
}

func (limiter *RateLimiter) prepareState() {
	limiter.bucket = limiter.JobsPerMin

	var f = 60 / float64(limiter.JobsPerMin)

	duration := time.Millisecond * time.Duration(f*1000)

	log.Printf("Job will be execute every %v\n", duration)

	limiter.ticker = time.NewTicker(duration)

	limiter.internalJobsChannel, limiter.safeClose = makeJobsChanelWithSafeCloser()

	limiter.jobQueue = make([]func(), 0)

	limiter.prepareWorkers()
}

func (limiter *RateLimiter) checkChannels() (error, bool) {
	select {
	case command := <-limiter.Command:
		if command == StopAndJoin {
			limiter.safeClose()
			log.Println("'Stop command' appears. Await all goroutins...")
			limiter.awaiter.Wait()
			log.Println("All goroutines stopped")
			return nil, true
		}
		if command == StopAndDetach {
			log.Println("'Stop command' appears. Shutdown without jobs awaiting")
			return nil, true
		}
	case <-limiter.ticker.C:
		if limiter.bucket < limiter.JobsPerMin {
			limiter.bucket++
			log.Printf("Increase bucket size (%d)\n", limiter.bucket)
		}
	case newJobs := <-limiter.jobQueueChannel:
		limiter.addNewJobs(newJobs)
	default:
		if limiter.bucket > 0 && len(limiter.jobQueue) > 0 {
			select {
			case limiter.internalJobsChannel <- limiter.getRandomJob():
				limiter.bucket--
				log.Printf("Job created, bucket size decreased (%d)\n", limiter.bucket)
			default:
			}
		}
	}

	return nil, false
}

func (limiter *RateLimiter) shuffleJobs() {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	for n := len(limiter.jobQueue); n > 0; n-- {
		randIndex := r.Intn(n)
		limiter.jobQueue[n-1], limiter.jobQueue[randIndex] =
			limiter.jobQueue[randIndex], limiter.jobQueue[n-1]
	}
}

func (limiter *RateLimiter) getRandomJob() func() {
	if limiter.nextJobIndex == len(limiter.jobQueue) {
		limiter.nextJobIndex = 0
		limiter.shuffleJobs()
	}

	job := limiter.jobQueue[limiter.nextJobIndex]

	limiter.nextJobIndex++

	return job
}

func (limiter *RateLimiter) addNewJobs(jobs []func()) {
	limiter.jobQueue = append(limiter.jobQueue, jobs...)
}

func (limiter *RateLimiter) prepareWorkers() {
	for i := 0; i < limiter.WorkersNumber; i++ {
		w := &worker{
			JobsChanel: limiter.internalJobsChannel,
		}

		go w.run(&limiter.awaiter)
	}
}

func makeJobsChanelWithSafeCloser() (jobChannel chan func(), safeClose func()) {
	jobChannel = make(chan func())

	var once sync.Once

	safeClose = func() {
		once.Do(func() {
			close(jobChannel)
		})
	}

	return
}
