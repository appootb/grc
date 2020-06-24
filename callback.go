package grc

import (
	"reflect"
	"sync"
)

type UpdateEvent func()

type CallbackFunc struct {
	Val DynamicType
	Evt UpdateEvent
}

type Callback interface {
	RegChan() chan<- *CallbackFunc
	EvtChan() chan<- DynamicType
}

var (
	callbackMgr = newCallback()
)

type callback struct {
	sync.RWMutex
	events map[DynamicType][]UpdateEvent

	evt chan DynamicType
	fn  chan *CallbackFunc
}

func newCallback() Callback {
	c := &callback{
		evt:    make(chan DynamicType, 50),
		fn:     make(chan *CallbackFunc, 50),
		events: make(map[DynamicType][]UpdateEvent),
	}
	go c.operate()
	return c
}

func (c *callback) RegChan() chan<- *CallbackFunc {
	return c.fn
}

func (c *callback) EvtChan() chan<- DynamicType {
	return c.evt
}

func (c *callback) operate() {
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
