package workerpool

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/remerge/cue"
)

type Pool struct {
	name     string
	callback WorkerCallback
	workers  []*Worker
	wg       sync.WaitGroup
	log      cue.Logger
}

func NewPool(name string, numWorkers int, callback WorkerCallback) *Pool {
	return &Pool{
		name:     name,
		callback: callback,
		workers:  make([]*Worker, numWorkers),
		log:      cue.NewLogger(name),
	}
}

// Send a message to one of the workers, determined by the provided id
func (p *Pool) Send(id int, msg interface{}) {
	if id < 0 {
		id = -id
	}
	p.workers[id%cap(p.workers)].Channel() <- msg
}

func (p *Pool) Recv() interface{} {
	cases := make([]reflect.SelectCase, cap(p.workers))

	for i, worker := range p.workers {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(worker.Channel()),
		}
	}

	_, value, ok := reflect.Select(cases)

	if !ok {
		return nil
	}

	return value.Interface()
}

func (p *Pool) Run() {
	p.log.WithFields(cue.Fields{
		"numWorkers": cap(p.workers),
	}).Info("worker pool spawn")

	for i := 0; i < cap(p.workers); i++ {
		p.workers[i] = NewWorker(fmt.Sprintf("%v/%v", p.name, i), p.callback)
		go func(num int) {
			p.wg.Add(1)
			defer p.wg.Done()
			p.workers[num].Run()
		}(i)
	}
}

func (p *Pool) Close() {
	p.log.Info("worker pool shutdown")
	for _, w := range p.workers {
		w.CloseWait()
	}
	p.log.Info("waiting for workers to shutdown")
	p.wg.Wait()
}
