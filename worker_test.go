package workerpool

import "testing"

var workerFunc = func(w *Worker) {
	for {
		w.Log.Debug("worker select")
		select {
		case <-w.Closer():
			w.Log.Debug("closer triggered")
			w.Done()
			w.Log.Debug("worker done")
			return
		case w.Channel() <- true:
			w.Log.Debug("worker msg sent")
			continue
		}
	}
}

func TestWorker(t *testing.T) {
	worker := NewWorker("test_worker", workerFunc)

	go worker.Run()

	msg := <-worker.Channel()
	if msg.(bool) != true {
		t.Error("Expected true, got ", msg)
	}

	worker.CloseWait()
}
