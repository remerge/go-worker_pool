package workerpool

import (
	"fmt"
	"sync"

	"github.com/remerge/cue"
)

type Pool struct {
	name     string
	callback WorkerCallback
	workers  []*Worker
	wg       sync.WaitGroup
	log      cue.Logger
	channel  chan interface{}
}

func NewPool(name string, numWorkers int, callback WorkerCallback) *Pool {
	return &Pool{
		name:     name,
		callback: callback,
		channel:  make(chan interface{}, numWorkers*2),
		workers:  make([]*Worker, numWorkers),
		log:      cue.NewLogger(name),
	}
}

// Send a message to one of the workers, determined by the provided id
func (p *Pool) Send(msg interface{}) {
	p.channel <- msg
}

func (p *Pool) Channel() chan interface{} {
	return p.channel
}

func (p *Pool) Run() {
	p.log.WithFields(cue.Fields{
		"numWorkers": cap(p.workers),
	}).Info("worker pool spawn")

	for i := 0; i < cap(p.workers); i++ {
		p.workers[i] = NewWorker(fmt.Sprintf("%v/%v", p.name, i), p.callback)
		p.wg.Add(1)
		go func(num int) {
			defer p.wg.Done()
			p.workers[num].Run()
		}(i)
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Close() {
	p.log.Info("worker pool shutdown")
	for _, w := range p.workers {
		w.CloseWait()
	}
	p.log.Info("waiting for workers to shutdown")
	p.wg.Wait()
}
