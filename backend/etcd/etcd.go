package etcd

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/appootb/grc/backend"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type Etcd struct {
	ctx context.Context
	*clientv3.Client
}

func NewProvider(ctx context.Context, endPoint, user, password string) (backend.Provider, error) {
	endPoints := strings.Split(endPoint, ",")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endPoints,
		DialTimeout: backend.DialTimeout,
		Username:    user,
		Password:    password,
	})
	if err != nil {
		return nil, err
	}
	return &Etcd{
		ctx:    ctx,
		Client: cli,
	}, nil
}

// Return provider type
func (p *Etcd) Type() string {
	return backend.Etcd
}

// Set value with the specified key
func (p *Etcd) Set(key, value string, ttl time.Duration) error {
	var options []clientv3.OpOption
	if ttl > 0 {
		ctx, cancel := context.WithTimeout(p.ctx, backend.WriteTimeout)
		lease, err := p.Grant(ctx, int64(ttl.Seconds()))
		cancel()
		if err != nil {
			return err
		}
		options = append(options, clientv3.WithLease(lease.ID))
	}

	ctx, cancel := context.WithTimeout(p.ctx, backend.WriteTimeout)
	defer cancel()
	_, err := p.Client.Put(ctx, key, value, options...)
	return err
}

// Get value of the specified key or directory
func (p *Etcd) Get(key string, dir bool) (backend.KVPairs, error) {
	var options []clientv3.OpOption
	if dir {
		options = append(options, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithTimeout(p.ctx, backend.ReadTimeout)
	defer cancel()
	resp, err := p.Client.Get(ctx, key, options...)
	if err != nil {
		return nil, err
	}
	kvs := make(backend.KVPairs, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		kvs = append(kvs, &backend.KVPair{
			Key:   string(kv.Key),
			Value: string(kv.Value),
		})
	}
	return kvs, nil
}

// Delete the specified key or directory
func (p *Etcd) Delete(key string, dir bool) error {
	return nil
}

// Watch for changes of the specified key or directory
func (p *Etcd) Watch(key string, dir bool) backend.EventChan {
	var options []clientv3.OpOption
	if dir {
		options = append(options, clientv3.WithPrefix())
	}

	eventsChan := make(backend.EventChan, backend.DefaultChanLen)
	etcdChan := p.Client.Watch(p.ctx, key, options...)

	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return

			case resp := <-etcdChan:
				if err := resp.Err(); err != nil {
					log.Println("etcd watch failure", err)
					// reset
					time.Sleep(time.Second)
					etcdChan = p.Client.Watch(p.ctx, key, options...)
					eventsChan <- &backend.WatchEvent{
						Type: backend.Reset,
					}
					continue
				}
				for _, evt := range resp.Events {
					wEvent := &backend.WatchEvent{
						KVPair: backend.KVPair{
							Key:   string(evt.Kv.Key),
							Value: string(evt.Kv.Value),
						},
					}
					if evt.Type == mvccpb.PUT {
						wEvent.Type = backend.Put
					} else {
						wEvent.Type = backend.Delete
					}
					eventsChan <- wEvent
				}
			}
		}
	}()

	return eventsChan
}

// Set and update ttl for the specified key
func (p *Etcd) KeepAlive(key, value string, ttl time.Duration) error {
	// grant lease
	ctx, cancel := context.WithTimeout(p.ctx, backend.WriteTimeout)
	lease, err := p.Grant(ctx, int64(ttl.Seconds()))
	cancel()
	if err != nil {
		return err
	}

	// put value with lease
	ctx, cancel = context.WithTimeout(ctx, backend.WriteTimeout)
	_, err = p.Client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	cancel()
	if err != nil {
		return err
	}

	// keep alive to etcd
	ch, err := p.Client.KeepAlive(ctx, lease.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ch:
				// do nothing
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// Close the provider connection
func (p *Etcd) Close() error {
	return p.Client.Close()
}
