package hw06pipelineexecution

import "time"

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	var out Out

	go ClosePipeline(in, out, done)

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

func ClosePipeline(in In, out Out, done In) {
	for {
		select {
		case <-done:
			// Перехватываю очередь значений, чтобы она закрылась на стороне отправки
			// for {
			// 	select {
			// 	case <-in:
			// 	case <-out:
			// 	default:
			// 		return
			// 	}
			// }
			for v := range in {
				// fmt.Println("Take in", v)
				_ = &v
			}

			for v := range out {
				// fmt.Println("Take out", v)
				_ = &v
			}
			return
		default:
			time.Sleep(time.Millisecond * 1)
		}
	}
}
