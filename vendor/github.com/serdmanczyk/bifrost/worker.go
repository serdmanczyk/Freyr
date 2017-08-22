package bifrost

func newWorker(workQueue chan chan *job) *worker {
	return &worker{
		workQueue: workQueue,
		work:      make(chan *job),
		stop:      make(chan bool),
	}
}

type worker struct {
	workQueue chan chan *job
	work      chan *job
	stop      chan bool
}

func (w *worker) start() {
	go func() {
		for {
			select {
			case w.workQueue <- w.work:
				job := <-w.work
				job.run()
			case <-w.stop:
				w.stop <- true
				close(w.stop)
				return
			}
		}
	}()
}
