package grc

import (
	"reflect"
	"sync"
)

var (
	CallbackMgr = NewCallback()
)

type UpdateEvent func()

type RegisterCallback struct {
	Val DynamicValue
	Evt UpdateEvent
}

type Callback struct {
	sync.RWMutex
	events map[DynamicValue][]UpdateEvent

	evt chan DynamicValue
	fn  chan *RegisterCallback
}

func NewCallback() *Callback {
	c := &Callback{
		evt:    make(chan DynamicValue, 50),
		fn:     make(chan *RegisterCallback, 50),
		events: make(map[DynamicValue][]UpdateEvent),
	}
	go c.operate()
	return c
}

func (c *Callback) RegChan() chan<- *RegisterCallback {
	return c.fn
}

func (c *Callback) EvtChan() chan<- DynamicValue {
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
