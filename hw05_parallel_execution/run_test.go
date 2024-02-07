package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	// "go.uber.org/goleak"
)

var runTime int64

func TestRun(t *testing.T) {
	// defer goleak.VerifyNone(t)
	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				time.Sleep(taskSleep)
				return nil
			})
		}

		workersCount := 10
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("m <= 0 should be considered as 'errors limit exceeded'", func(t *testing.T) {
		tasksCount := 10
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := -1
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.Equal(t, runTasksCount, int32(0), "no tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				time.Sleep(taskSleep)
				return nil
			})
		}

		workersCount := 10
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("tasks without errors (with require.Eventually", func(t *testing.T) {
		tasksCount := 10
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				start := time.Now()
				atomic.AddInt32(&runTasksCount, 1)
				time.Sleep(10 * time.Millisecond)
				elapsedTime := time.Since(start)
				atomic.AddInt64(&runTime, int64(elapsedTime))
				// fmt.Println("TASK: total running time in all goroutines: ", fmt.Sprint(time.Duration(runTime)))
				return nil
			})
		}

		workersCount := 10
		maxErrorsCount := 1
		require.Eventually(t, func() bool {
			_ = Run(tasks, workersCount, maxErrorsCount)
			fmt.Println("Total running time in all goroutines:", time.Duration(runTime))
			return true
		}, 50*time.Millisecond, time.Millisecond, "duration:"+fmt.Sprint(time.Duration(runTime))+
			", tasks were run sequentially?")
	})
}
