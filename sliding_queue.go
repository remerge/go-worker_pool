package workerpool

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
)

// SlidingQueue guarantees sequential results processing
type SlidingQueue struct {
	ctx    context.Context
	cancel context.CancelFunc

	lastIndex  int64 // index clock
	lastCommit int64
	available  chan int64

	mu            sync.Mutex
	slotsInFlight map[int64]*slidingQueueSlot
}

// NewSlidingQueue returns new sliding queue with given capacity
func NewSlidingQueue(capacity int) (q *SlidingQueue) {
	q = &SlidingQueue{
		// lastCommit:    -1,
		available:     make(chan int64, capacity),
		slotsInFlight: map[int64]*slidingQueueSlot{},
	}
	q.ctx, q.cancel = context.WithCancel(context.Background())
	for i := 0; i < capacity; i++ {
		q.addAvailable()
	}
	go func() {
		<-q.ctx.Done()
		q.mu.Lock()
		defer q.mu.Unlock()
		q.evaluateSlots()
	}()
	return q
}

// Close closes SlidingQueue and cancels context of all slots
func (q *SlidingQueue) Close() error {
	q.cancel()
	return nil
}

func (q *SlidingQueue) Acquire(ctx context.Context, fn func()) (commit func(), err error) {
	select {
	case <-q.ctx.Done():
		return nil, q.ctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	case index := <-q.available:
		q.mu.Lock()
		q.slotsInFlight[index] = &slidingQueueSlot{
			commit: fn,
		}
		q.mu.Unlock()
		return func() {
			q.onCommit(index)
		}, nil
	}
}

func (q *SlidingQueue) addAvailable() {
	select {
	case <-q.ctx.Done():
	case q.available <- atomic.AddInt64(&q.lastIndex, 1):
	}
}

func (q *SlidingQueue) onCommit(index int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.slotsInFlight[index].done = true
	q.evaluateSlots()
	q.addAvailable()
}

func (q *SlidingQueue) evaluateSlots() {
	var done []int64
	for i, slot := range q.slotsInFlight {
		if slot.done {
			done = append(done, i)
		}
	}
	sort.Slice(done, func(i, j int) bool {
		return done[i] < done[j]
	})

	if len(done) > 0 {
		for _, i := range done {
			if i-q.lastCommit > 1 {
				break
			}
			q.slotsInFlight[i].commit()
			q.lastCommit = i
			delete(q.slotsInFlight, i)
		}
	}
}

type slidingQueueSlot struct {
	commit func()
	done   bool
}
