package workerpool

import (
	"sync"

	"github.com/bobziuchkovski/cue"
)

type Pool interface {
	Run()
}

type pool struct {
	callback WorkerCallback
	workers  []*Worker
	wg       sync.WaitGroup
	log      cue.Logger
}

func NewPool(numWorkers int, callback WorkerCallback) Pool {
	return &pool{
		callback: callback,
		workers:  make([]*Worker, numWorkers),
		log:      cue.NewLogger("workerpool"),
	}
}

func (p *pool) Run() {
	p.log.WithFields(cue.Fields{
		"numWorkers": cap(p.workers),
	}).Info("worker pool spawn")

	for i := 0; i < cap(p.workers); i++ {
		go func(num int) {
			p.wg.Add(1)
			defer p.wg.Done()
			p.workers[num] = NewWorker(p.callback)
			p.workers[num].Run()
		}(i)
	}
}

func (p *pool) Close() {
	p.log.Info("worker pool shutdown")
	for _, w := range p.workers {
		go w.WaitClose()
	}
	p.wg.Wait()
}
