package grc

import (
	"reflect"
	"sync"
)

var (
	CallbackMgr = NewCallback()
)

type RegisterCallback struct {
	Val AtomicUpdate
	Evt UpdateEvent
}

type Callback struct {
	sync.RWMutex
	events map[AtomicUpdate][]UpdateEvent

	evt chan AtomicUpdate
	fn  chan *RegisterCallback
}

func NewCallback() *Callback {
	c := &Callback{
		evt:    make(chan AtomicUpdate, 50),
		fn:     make(chan *RegisterCallback, 50),
		events: make(map[AtomicUpdate][]UpdateEvent),
	}
	go c.operate()
	return c
}

func (c *Callback) RegChan() chan<- *RegisterCallback {
	return c.fn
}

func (c *Callback) EvtChan() chan<- AtomicUpdate {
	return c.evt
}

func (c *Callback) operate() {
	for {
		select {
		case fn := <-c.fn:
			if reflect.ValueOf(fn.Val).IsNil() {
				panic("grc: cannot register callback of a pointer type.")
			}
			c.Lock()
			c.events[fn.Val] = append(c.events[fn.Val], fn.Evt)
			c.Unlock()

		case val := <-c.evt:
			c.RLock()
			if events, ok := c.events[val]; ok {
				for _, evt := range events {
					evt()
				}
			}
			c.RUnlock()
		}
	}
}
