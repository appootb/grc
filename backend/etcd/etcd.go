package etcd

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/appootb/grc/backend"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type Etcd struct {
	ctx context.Context
	*clientv3.Client
}

func NewProvider(ctx context.Context, endPoints []string, username, password string) (backend.Provider, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            endPoints,
		DialTimeout:          backend.DialTimeout,
		DialKeepAliveTime:    backend.KeepAliveTime,
		DialKeepAliveTimeout: backend.DialTimeout,
		Username:             username,
		Password:             password,
	})
	if err != nil {
		return nil, err
	}
	return &Etcd{
		ctx:    ctx,
		Client: cli,
	}, nil
}

// Type returns the provider type.
func (p *Etcd) Type() string {
	return backend.Etcd
}

// Set value for the specified key with a specified ttl.
func (p *Etcd) Set(key, value string, ttl time.Duration) error {
	var options []clientv3.OpOption
	if ttl > 0 {
		leaseCtx, leaseCancel := context.WithTimeout(p.ctx, backend.WriteTimeout)
		defer leaseCancel()
		lease, err := p.Grant(leaseCtx, int64(ttl.Seconds()))
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

// Get the value of the specified key or directory.
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

// Incr invokes an atomic value increase for the specified key.
func (p *Etcd) Incr(key string) (int64, error) {
	session, err := concurrency.NewSession(p.Client)
	if err != nil {
		return 0, err
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, key)
	ctx, cancel := context.WithTimeout(p.ctx, backend.WriteTimeout*2)
	defer cancel()
	if err = mutex.Lock(ctx); err != nil {
		return 0, err
	}
	defer mutex.Unlock(p.ctx)

	num := int64(0)
	kvs, err := p.Get(key, false)
	if err != nil {
		return 0, err
	} else if len(kvs) > 0 {
		num, _ = strconv.ParseInt(kvs[0].Value, 10, 64)
	}
	num++
	if err = p.Set(key, strconv.FormatInt(num, 10), 0); err != nil {
		return 0, err
	}
	return num, nil
}

// Delete the specified key or directory.
func (p *Etcd) Delete(key string, dir bool) error {
	var options []clientv3.OpOption
	if dir {
		options = append(options, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithTimeout(p.ctx, backend.WriteTimeout)
	defer cancel()
	_, err := p.Client.Delete(ctx, key, options...)
	return err
}

// Watch for changes of the specified key or directory.
func (p *Etcd) Watch(key string, dir bool) (backend.EventChan, error) {
	revision, err := p.sync(key, dir, nil)
	if err != nil {
		return nil, err
	}
	//
	eventsChan := make(backend.EventChan, backend.DefaultChanLen)
	//
	go p.watch(key, dir, revision, eventsChan)

	return eventsChan, nil
}

func (p *Etcd) sync(key string, dir bool, eventsChan backend.EventChan) (int64, error) {
	var options []clientv3.OpOption
	if dir {
		options = append(options, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithTimeout(p.ctx, backend.ReadTimeout)
	defer cancel()
	//
	resp, err := p.Client.Get(ctx, key, options...)
	if err != nil {
		return 0, err
	}
	//
	if eventsChan != nil {
		for _, kv := range resp.Kvs {
			eventsChan <- &backend.WatchEvent{
				Type: backend.Reset,
				KVPair: backend.KVPair{
					Key:   string(kv.Key),
					Value: string(kv.Value),
				},
			}
		}
	}
	//
	return resp.Header.Revision, nil
}

func (p *Etcd) watch(key string, dir bool, revision int64, eventsChan backend.EventChan) {
Retry:
	options := []clientv3.OpOption{
		clientv3.WithRev(revision),
		clientv3.WithProgressNotify(),
	}
	if dir {
		options = append(options, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithCancel(p.ctx)
	etcdChan := p.Client.Watch(ctx, key, options...)

	for {
		select {
		case <-p.ctx.Done():
			cancel()
			return

		case resp := <-etcdChan:
			if resp.CompactRevision > 0 {
				time.Sleep(time.Second)
				log.Println("grc: etcd revision compacted")
				revision, _ = p.sync(key, dir, eventsChan)
				goto Retry
			} else if err := resp.Err(); err != nil {
				cancel()
				time.Sleep(time.Second * 5)
				log.Println("grc: etcd watch error, ", err.Error())
				goto Retry
			}
			//
			if resp.Header.Revision > 0 {
				revision = resp.Header.Revision
			}
			if resp.IsProgressNotify() {
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
			//
			revision = resp.Header.GetRevision()
		}
	}
}

// KeepAlive sets value and updates the ttl for the specified key.
func (p *Etcd) KeepAlive(key, value string, ttl time.Duration) error {
	ch, err := p.keepAlive(key, value, ttl, false)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case m := <-ch:
				// channel closed, retry
				if m == nil {
					ch, _ = p.keepAlive(key, value, ttl, true)
				}
			case <-p.ctx.Done():
				_, err = p.Client.Delete(context.TODO(), key)
				if err != nil {
					log.Println("grc: etcd KeepAlive stopping, ", err.Error())
				}
				return
			}
		}
	}()
	return nil
}

// Close the provider connection.
func (p *Etcd) Close() error {
	return p.Client.Close()
}

func (p *Etcd) keepAlive(key, value string, ttl time.Duration, withRetry bool) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
Retry:
	// grant lease
	ctx, cancel := context.WithTimeout(p.ctx, backend.WriteTimeout)
	lease, err := p.Grant(ctx, int64(ttl.Seconds()))
	cancel()
	if err != nil {
		if withRetry {
			time.Sleep(backend.RetryTimeout)
			goto Retry
		}
		return nil, err
	}

	// put value with lease
	ctx, cancel = context.WithTimeout(p.ctx, backend.WriteTimeout)
	_, err = p.Client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	cancel()
	if err != nil {
		if withRetry {
			time.Sleep(backend.RetryTimeout)
			goto Retry
		}
		return nil, err
	}

	// keep alive to etcd
	ch, err := p.Client.KeepAlive(p.ctx, lease.ID)
	if err != nil {
		if withRetry {
			time.Sleep(backend.RetryTimeout)
			goto Retry
		}
		return nil, err
	}

	return ch, nil
}
