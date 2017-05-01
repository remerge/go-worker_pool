package workerpool

type WorkerCallback func(*Worker)

type Worker struct {
	closer   chan bool
	done     chan bool
	callback WorkerCallback
}

func NewWorker(callback WorkerCallback) *Worker {
	return &Worker{
		closer:   make(chan bool),
		callback: callback,
	}
}

func (w *Worker) Run() {
	w.callback(w)
}

func (w *Worker) Closer() chan bool {
	return w.closer
}

func (w *Worker) WaitClose() {
	close(w.closer)
	<-w.done
}

func (w *Worker) Done() {
	close(w.done)
}
