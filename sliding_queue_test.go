package workerpool

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlidingQueque_Acquire(t *testing.T) {
	intSlice := func(n int) []int {
		s := make([]int, n)
		for i := 0; i < n; i++ {
			s[i] = i
		}
		return s
	}

	stopChan := make(chan struct{})
	var results []int
	var mu sync.Mutex

	q := NewSlidingQueue(32)
	for i := 0; i < 100; i++ {
		cb := func(i int) func() {
			return func() {
				mu.Lock()
				defer mu.Unlock()
				results = append(results, i)
				if len(results) == 100 {
					close(stopChan)
				}
			}
		}
		commit, err := q.Acquire(context.Background(), cb(i))
		assert.NoError(t, err)
		go func(i int, commit func()) {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(100)))
			commit()
		}(i, commit)
	}
	<-stopChan
	_ = q.Close()
	assert.Equal(t, intSlice(100), results)
}
