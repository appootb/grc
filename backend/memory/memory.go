package memory

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/appootb/grc/backend"
)

var (
	zeroTime = time.Unix(0, 0)
)

type node struct {
	k      string
	v      string
	expire time.Time
}

type watch struct {
	ch     backend.EventChan
	key    string
	prefix bool
}

type Memory struct {
	kvs map[string]*node
	ws  []*watch

	event  backend.EventChan
	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
}

func NewProvider() backend.Provider {
	p := &Memory{
		kvs:   make(map[string]*node),
		event: make(backend.EventChan, 10),
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	go p.checkTTL()
	go p.checkWatch()
	return p
}

// Type returns the provider type.
func (p *Memory) Type() string {
	return backend.Memory
}

// Set value for the specified key with a specified ttl.
func (p *Memory) Set(key, value string, ttl time.Duration) error {
	expire := time.Now().Add(ttl)
	if ttl == 0 {
		expire = zeroTime
	}
	p.Lock()
	p.kvs[key] = &node{
		k:      key,
		v:      value,
		expire: expire,
	}
	p.Unlock()
	p.event <- &backend.WatchEvent{
		Type: backend.Put,
		KVPair: backend.KVPair{
			Key:   key,
			Value: value,
		},
	}
	return nil
}

// Get the value of the specified key or directory.
func (p *Memory) Get(key string, dir bool) (backend.KVPairs, error) {
	p.RLock()
	defer p.RUnlock()
	if !dir {
		if n, ok := p.kvs[key]; !ok {
			return backend.KVPairs{}, nil
		} else {
			return backend.KVPairs{
				{
					Key:   key,
					Value: n.v,
				},
			}, nil
		}
	}
	//
	var kvs backend.KVPairs
	for k, v := range p.kvs {
		if strings.HasPrefix(k, key) {
			kvs = append(kvs, &backend.KVPair{
				Key:   k,
				Value: v.v,
			})
		}
	}
	return kvs, nil
}

// Incr invokes an atomic value increase for the specified key.
func (p *Memory) Incr(key string) (int64, error) {
	p.Lock()
	defer p.Unlock()
	n, ok := p.kvs[key]
	if !ok {
		n = &node{
			k:      key,
			v:      "0",
			expire: zeroTime,
		}
	}
	v, _ := strconv.ParseInt(n.v, 10, 64)
	v++
	n.v = strconv.FormatInt(v, 10)
	p.kvs[key] = n
	return v, nil
}

// Delete the specified key or directory.
func (p *Memory) Delete(key string, dir bool) error {
	p.Lock()
	defer p.Unlock()
	if !dir {
		delete(p.kvs, key)
		return nil
	}
	//
	for k := range p.kvs {
		if strings.HasPrefix(k, key) {
			delete(p.kvs, k)
		}
	}
	return nil
}

// Watch for changes of the specified key or directory.
func (p *Memory) Watch(key string, dir bool) (backend.EventChan, error) {
	p.Lock()
	defer p.Unlock()
	ch := make(backend.EventChan, 10)
	p.ws = append(p.ws, &watch{
		ch:     ch,
		key:    key,
		prefix: dir,
	})
	return ch, nil
}

// KeepAlive sets value and updates the ttl for the specified key.
func (p *Memory) KeepAlive(key, value string, ttl time.Duration) error {
	return p.Set(key, value, 0)
}

// Close the provider connection.
func (p *Memory) Close() error {
	p.cancel()
	return nil
}

func (p *Memory) checkTTL() {
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-p.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			p.Lock()
			for k, v := range p.kvs {
				if v.expire.Sub(zeroTime) > 0 && time.Now().Sub(v.expire) > 0 {
					delete(p.kvs, k)
					p.event <- &backend.WatchEvent{
						Type: backend.Delete,
						KVPair: backend.KVPair{
							Key:   v.k,
							Value: v.v,
						},
					}
				}
			}
			p.Unlock()
		}
	}
}

func (p *Memory) checkWatch() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case evt := <-p.event:
			p.RLock()
			for _, w := range p.ws {
				if evt.Key == w.key ||
					w.prefix && strings.HasPrefix(evt.Key, w.key) {
					w.ch <- evt
				}
			}
			p.RUnlock()
		}
	}
}
