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
		if msg.(bool) != true {
			t.Error("Expected true, got ", msg)
		}
	}

	testLog.Info("pool close")
	pool.Close()
}
