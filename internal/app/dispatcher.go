package app

type Collector struct {
	Job chan Job
	end chan bool
}

var workerChannel = make(chan chan Job)

func (di *Diun) StartDispatcher(workerCount int) Collector {
	var i int
	var workers []worker
	input := make(chan Job)
	end := make(chan bool)
	collector := Collector{
		Job: input,
		end: end,
	}

	for i < workerCount {
		i++
		worker := worker{
			id:         i,
			diun:       di,
			workerPool: workerChannel,
			jobChannel: make(chan Job),
			end:        make(chan bool),
		}
		worker.Start()
		workers = append(workers, worker)
	}

	go func() {
		for {
			select {
			case <-end:
				for _, w := range workers {
					w.Stop()
				}
				return
			case work := <-input:
				worker := <-workerChannel
				worker <- work
			}
		}
	}()

	return collector
}
