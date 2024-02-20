package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	var out Out

	for i, stage := range stages {
		i := i
		if i == 0 {
			out = stage(in)
		} else {
			out = stage(out)
		}
	}

	return CheckDone(out, done)
}

func CheckDone(in In, done In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case data, ok := <-in:
				if !ok {
					return
				}
				out <- data
			}
		}
	}()

	return out
}
