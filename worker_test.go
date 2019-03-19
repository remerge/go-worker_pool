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
	if !msg.(bool) {
		t.Error("Expected true, got ", msg)
	}

	worker.CloseWait()
}

var readerFunc = func(w *Worker) {
	for {
		select {
		case <-w.Closer():
			w.Done()
			return
		case msg, ok := <-w.Channel():
			if !ok {
				panic("channel closed unexpectedly")
			}
			if msg != "testmessage" {
				panic("message was different from testmessage")
			}
			testLog.Debug(msg.(string))
		}
	}
}

func TestWorker_Send(_ *testing.T) {
	reader := NewWorker("test_reader", readerFunc)

	go reader.Run()

	reader.Channel() <- "testmessage"

	reader.Close()
}
