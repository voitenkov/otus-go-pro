package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	mu                                            sync.Mutex
	cond                                          *sync.Cond
	wg                                            *sync.WaitGroup
	ErrErrorsLimitExceeded                        = errors.New("errors limit exceeded")
	errorsCount, runTasksCount, taskNo, tasksSent int32
	tasksCount                                    int
)

type Task func() error

func producer(tasksSlice []Task, tasksChan chan<- Task, syncChan <-chan struct{}, doneChan chan<- struct{}, m int) {
	defer wg.Done()
	mu.Lock()
	fmt.Println("PRODUCER: started service")
	for _, t := range tasksSlice {
		tasksChan <- t
		atomic.AddInt32(&tasksSent, 1)
	}
	close(tasksChan)
	mu.Unlock()
	cond.Broadcast()

	tasksCount = len(tasksSlice)
	// fmt.Printf("PRODUCER: errors count: %d, errors limit: %d, tasks count: %d, tasks total: %d\n",
	// 	errorsCount, m, runTasksCount, tasksCount)
	for (int(errorsCount) < m) && (int(runTasksCount) < tasksCount) {
		<-syncChan
		// fmt.Printf("PRODUCER: errors count: %d, errors limit: %d, tasks count: %d, tasks total: %d\n",
		// 	errorsCount, m, runTasksCount, tasksCount)
	}

	close(doneChan)
	// fmt.Printf("PRODUCER: errors count: %d, errors limit: %d, tasks count: %d, tasks total: %d\n",
	// 	errorsCount, m, runTasksCount, tasksCount)
	fmt.Println("PRODUCER: stopped service")
}

func worker(id int, tasksChan <-chan Task, syncChan chan<- struct{}, doneChan <-chan struct{}, m int) {
	defer wg.Done()
	mu.Lock()
	for int(tasksSent) < tasksCount {
		cond.Wait()
	}
	mu.Unlock()
	workerErrorsCount := 0
	var err error
	for task := range tasksChan {
		// fmt.Printf("WORKER %d: errors count: %d, errors limit: %d, tasks count: %d, tasks total: %d\n",
		// 	id, errorsCount, m, runTasksCount, tasksCount)
		if int(errorsCount) < m {
			atomic.AddInt32(&taskNo, 1)
			// currentTask := int(taskNo)
			// fmt.Printf("WORKER %d: started task %d\n", id, currentTask)
			err = task()
			// fmt.Printf("WORKER %d: finished task %d\n", id, currentTask)
			if err != nil {
				atomic.AddInt32(&errorsCount, 1)
				workerErrorsCount++
			}
			atomic.AddInt32(&runTasksCount, 1)
			select {
			case <-doneChan:
				// fmt.Printf("WORKER %d: stopped with %d errors\n", id, workerErrorsCount)
				return
			default:
				syncChan <- struct{}{}
			}
		} else {
			<-doneChan
			// fmt.Printf("WORKER %d: stopped with %d errors\n", id, workerErrorsCount)
			return
		}
	}
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasksSlice []Task, n, m int) error {
	fmt.Println("-----------------------")
	fmt.Println("RUNNER: started service")
	var err error
	tasksCount := len(tasksSlice)
	errorsCount = 0
	runTasksCount = 0
	if m <= 0 {
		fmt.Printf("RUNNER: errors count: %d, errors limit: %d, tasks count: %d, tasks total: %d\n",
			errorsCount, m, runTasksCount, tasksCount)
		fmt.Println(ErrErrorsLimitExceeded)
		fmt.Println("RUNNER: stopped service")
		return ErrErrorsLimitExceeded
	}
	tasksChan := make(chan Task, tasksCount)
	syncChan := make(chan struct{}, n)
	doneChan := make(chan struct{})
	taskNo = 0
	wg = &sync.WaitGroup{}
	cond = &sync.Cond{}
	wg.Add(1)

	// Create producer goroutine
	go producer(tasksSlice, tasksChan, syncChan, doneChan, m)

	// Create worker goroutines
	for w := 1; w <= n; w++ {
		wg.Add(1)
		go worker(w, tasksChan, syncChan, doneChan, m)
	}

	wg.Wait()
	close(syncChan)
	fmt.Printf("RUNNER: errors count: %d, errors limit: %d, tasks count: %d, tasks total: %d\n",
		errorsCount, m, runTasksCount, tasksCount)
	if int(errorsCount) >= m {
		err = ErrErrorsLimitExceeded
		fmt.Println(ErrErrorsLimitExceeded)
	}
	fmt.Println("RUNNER: stopped service")
	return err
}
