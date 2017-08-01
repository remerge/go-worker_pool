package workerpool

import (
	"testing"
)

func TestPool(t *testing.T) {
	numWorkers := 4

	pool := NewPool("test_pool", numWorkers, workerFunc)

	pool.Run()

	for i := 0; i < numWorkers; i++ {
		testLog.Infof("recv msg %v", i)
		msg := pool.Recv()
		if !msg.(bool) {
			t.Error("Expected true, got ", msg)
		}
	}

	testLog.Info("pool close")
	pool.Close()
}

func TestPool_Send(t *testing.T) {
	numWorkers := 4

	pool := NewPool("test_pool", numWorkers, readerFunc)

	pool.Run()

	for i := 0; i < 10; i++ {
		pool.Send(i, "testmessage")
	}
	for i := -1; i > -10; i-- {
		pool.Send(i, "testmessage")
	}

	pool.Close()
}
