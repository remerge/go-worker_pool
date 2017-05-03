package workerpool

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/bobziuchkovski/cue"
)

type Pool interface {
	Run()
	Close()
	Send(id int, msg interface{})
	Recv() interface{}
}

type pool struct {
	callback WorkerCallback
	workers  []*Worker
	wg       sync.WaitGroup
	log      cue.Logger
}

func NewPool(numWorkers int, callback WorkerCallback) Pool {
	p := &pool{}
	p.callback = callback
	p.workers = make([]*Worker, numWorkers)
	p.log = cue.NewLogger(fmt.Sprintf("workerpool-%p", p))
	return p
}

func (p *pool) Send(id int, msg interface{}) {
	p.log.WithFields(cue.Fields{
		"id":  id,
		"msg": msg,
	}).Debug("send")
	p.workers[id%cap(p.workers)].channel <- msg
	p.log.WithFields(cue.Fields{
		"id":  id,
		"msg": msg,
	}).Debug("sent")
}

func (p *pool) Recv() interface{} {
	cases := make([]reflect.SelectCase, cap(p.workers))
	for i, worker := range p.workers {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(worker.Channel()),
		}
	}

	chosen, value, ok := reflect.Select(cases)
	p.log.WithFields(cue.Fields{
		"chosen": chosen,
		"value":  value,
		"ok":     ok,
	}).Debug("select")

	return value.Interface()
}

func (p *pool) Run() {
	p.log.WithFields(cue.Fields{
		"numWorkers": cap(p.workers),
	}).Info("worker pool spawn")

	for i := 0; i < cap(p.workers); i++ {
		p.workers[i] = NewWorker(p.callback)
		go func(num int) {
			p.wg.Add(1)
			defer p.wg.Done()
			p.log.Debugf("run worker %v", num)
			p.workers[num].Run()
		}(i)
	}
}

func (p *pool) Close() {
	p.log.Info("worker pool shutdown")
	for _, w := range p.workers {
		w.CloseWait()
	}
	p.log.Info("waiting for workers to shutdown")
	p.wg.Wait()
}
