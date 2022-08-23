package workerpool

import (
	"sync"

	"github.com/remerge/cue"
)

type WorkerCallback func(*Worker)

type Worker struct {
	sync.Mutex
	Log         cue.Logger
	notifyClose chan bool
	notifyDone  chan bool
	closed      bool
	done        bool
	callback    WorkerCallback
}

func NewWorker(name string, callback WorkerCallback) (w *Worker) {
	w = &Worker{}
	w.Log = cue.NewLogger(name)
	w.notifyClose = make(chan bool)
	w.notifyDone = make(chan bool)
	w.callback = callback
	return w
}

func (w *Worker) Run() {
	w.Log.Debug("run loop start callback")
	w.callback(w)
}

func (w *Worker) Closer() chan bool {
	return w.notifyClose
}

func (w *Worker) Close() {
	w.Lock()
	defer w.Unlock()
	if !w.closed {
		w.Log.Debug("run loop notify close")
		close(w.notifyClose)
		w.closed = true
	}
}

func (w *Worker) Wait() {
	w.Log.Debug("run loop wait")
	<-w.notifyDone
	w.Log.Debug("run loop done")
}

func (w *Worker) CloseWait() {
	w.Close()
	w.Wait()
}

func (w *Worker) Done() {
	// worker loop might have called Done() without Close() being called
	// before, so let's close the channel for good measure
	w.Close()

	w.Lock()
	defer w.Unlock()

	if !w.done {
		w.Log.Debug("run loop notify done")
		close(w.notifyDone)
		w.done = false
	}
}
