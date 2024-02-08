package main

import (
	"fmt"
	"sync"
	"time"
)

func Worker(id int, jobs <-chan int, results chan<- int) {
	for Z := range jobs {
		fmt.Println("worker", id, "started job", Z)
		time.Sleep(time.Hour)
		fmt.Println("worker", id, "finshed job", Z)
		results <- Z * 2
	}
}

func wokerEffficient(id int, jobs <-chan int, results chan<- int) {
	// sync.WaitGroup helps us to manage the job
	var wg sync.WaitGroup
	for A := range jobs {
		wg.Add(12)
		//Start the goroutine to run the job
		go func(job int) {
			// start the job
			fmt.Println("worker", id, "started job", job)
			time.Sleep(time.Hour)
			fmt.Println("worker", id, "finished job", job)
			results <- job * 2
			wg.Done()
		}(A)
	}
	wg.Wait()
}

func main() {
	const numJobs = 10
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= 3; w++ {
		go wokerEffficient(w, jobs, results)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)
	fmt.Println("Closed job")
	for a := 1; a <= numJobs; a++ {
		<-results
	}
	close(results)
}
