package pools

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type RateLimiter struct {
	WorkersNumber int
	Rpm           int
	Command       chan int
}

func (limiter *RateLimiter) Run() error {
	if limiter.Rpm <= 0 {
		return errors.New("rpm cannot be less than 1")
	}

	bucket := limiter.Rpm

	var f = 60 / float64(limiter.Rpm)

	duration := time.Millisecond * time.Duration(f*1000)

	fmt.Printf("Job will be execute every %v\n", duration)

	ticker := time.NewTicker(duration)

	ch, safeClose := makeChanelWithSafeCloser()

	defer safeClose()

	jobNumber := 0

	var awaiter sync.WaitGroup

	for i := 0; i < limiter.WorkersNumber; i++ {
		w := &Worker{
			Ch: ch,
			Job: func() {
				minDelay := 5
				maxDelay := 10

				time.Sleep(time.Duration(rand.Intn(maxDelay-minDelay)+minDelay) * time.Second)
			},
		}

		go w.Run(&awaiter)
	}

	for {
		select {
		case command := <-limiter.Command:
			if command == StopAndJoin {
				safeClose()
				fmt.Println("Stop command appears. Await all goroutins...")
				awaiter.Wait()
				fmt.Println("All goroutines stopped")
				return nil
			}
			if command == StopAndDetach {
				fmt.Println("Stop command appears. Shutdown without jobs awaiting")
				return nil
			}
		case <-ticker.C:
			if bucket < limiter.Rpm {
				bucket++
				fmt.Printf("Increase bucket size (%d)\n", bucket)
			}
		default:
			if bucket > 0 {
				select {
				case ch <- jobNumber:
					jobNumber++
					bucket--
					fmt.Printf("Job created, bucket size decriazed (%d)\n", bucket)
				default:
				}
			}
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func makeChanelWithSafeCloser() (chan int, func()) {
	ch := make(chan int)

	var once sync.Once

	return ch, func() {
		once.Do(func() {
			close(ch)
		})
	}
}
