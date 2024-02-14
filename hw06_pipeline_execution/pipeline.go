package hw06pipelineexecution

import (
	"fmt"
	"time"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in Bi, done In, stages ...Stage) Out {
	var out Out

	go ClosePipeline(in, done)

	for i, stage := range stages {
		i := i
		if i == 0 {
			out = stage(in)
		} else {
			out = stage(out)
		}
	}
	return out
}

func ClosePipeline(in Bi, done In) {
	for {
		select {
		case <-done:
			// Перехватываю очередь значений, чтобы она закрылась на стороне отправки
			for v := range in {
				fmt.Println("Take", v)
			}
			return
		default:
			time.Sleep(time.Millisecond * 1)
		}
	}
}
