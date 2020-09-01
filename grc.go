package grc

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/appootb/grc/backend"
)

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "grc: Config type(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "grc: Config type(non-pointer " + e.Type.String() + ")"
	}
	return "grc: Config type(nil " + e.Type.String() + ")"
}

type RemoteConfig struct {
	svc sync.Map
	ctx context.Context

	path         string
	autoCreation bool
	provider     backend.Provider
}

func New(opts ...Option) (*RemoteConfig, error) {
	rc := &RemoteConfig{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt.apply(rc)
	}

	basePath := backend.ServiceDiscoveryPrefixKey(rc.path)
	// Watch for service nodes updated.
	evtChan := rc.provider.Watch(basePath, true)
	// Get services.
	if err := rc.getServices(basePath); err != nil {
		return nil, err
	}
	go rc.watchServiceEvent(basePath, evtChan)
	return rc, nil
}

func (rc *RemoteConfig) RegisterNode(service, nodeAddr string, opts ...NodeOption) (int64, error) {
	node := &Node{
		TTL:      time.Second * 3,
		Service:  service,
		Address:  nodeAddr,
		Weight:   1,
		Metadata: map[string]string{},
	}
	for _, opt := range opts {
		opt(node)
	}

	if node.ops {
		err := rc.loadUniqueID(node)
		if err != nil {
			return 0, err
		}
	} else {
		uniqueID, err := rc.provider.Incr(backend.ServiceNodeIDIncrKey(rc.path, node.Service))
		if err != nil {
			return 0, err
		}
		node.UniqueID = uniqueID
	}

	key := backend.ServiceDiscoveryKey(rc.path, service, node.Address)
	if err := rc.provider.KeepAlive(key, node.String(), node.TTL); err != nil {
		return 0, err
	}
	return node.UniqueID, nil
}

func (rc *RemoteConfig) GetNodes(service string) Nodes {
	nodes, ok := rc.svc.Load(service)
	if !ok {
		return Nodes{}
	}
	return nodes.(Nodes)
}

func (rc *RemoteConfig) RegisterConfig(service string, v interface{}) error {
	cfg := reflect.ValueOf(v)
	if cfg.Kind() != reflect.Ptr || cfg.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(cfg)}
	}

	basePath := backend.ServiceConfigKey(rc.path, service)
	// Create/update default config value if not exist.
	if rc.autoCreation {
		err := rc.remoteConfigMigration(basePath, reflect.TypeOf(v))
		if err != nil {
			return err
		}
	}

	// Watch for config updated.
	evtChan := rc.provider.Watch(basePath, true)
	// Initialize the config.
	if err := rc.getConfig(basePath, configElem(cfg), false); err != nil {
		return err
	}
	go rc.watchConfigEvent(basePath, evtChan, configElem(cfg))
	return nil
}

func (rc *RemoteConfig) remoteConfigMigration(basePath string, t reflect.Type) error {
	kvs, err := rc.provider.Get(basePath, true)
	if err != nil {
		return err
	}
	reflectKVs := parseConfig(t, "").KVs(basePath)
	for _, kv := range kvs {
		delete(reflectKVs, kv.Key)
	}
	for k, v := range reflectKVs {
		err := rc.provider.Set(k, v, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *RemoteConfig) getConfig(basePath string, cfg reflect.Value, forUpdate bool) error {
	kvs, err := rc.provider.Get(basePath, true)
	if err != nil {
		return err
	}
	for _, pair := range kvs {
		err := rc.setConfig(basePath, pair, cfg, forUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *RemoteConfig) watchConfigEvent(basePath string, ch backend.EventChan, cfg reflect.Value) {
	var (
		err error
	)

	for {
		select {
		case <-rc.ctx.Done():
			return

		case evt := <-ch:
			if evt.Type == backend.Reset {
				err = rc.getConfig(basePath, cfg, true)
			} else {
				err = rc.setConfig(basePath, &evt.KVPair, cfg, true)
			}
			if err != nil {
				log.Println("grc: watchConfigEvent failed:", err.Error(), evt.Type, evt.Key)
			}
		}
	}
}

func (rc *RemoteConfig) setConfig(basePath string, pair *backend.KVPair, cfg reflect.Value, forUpdate bool) error {
	var item backend.ConfigItem
	if err := json.Unmarshal([]byte(pair.Value), &item); err != nil {
		return err
	}

	fieldPath := strings.Split(strings.TrimPrefix(pair.Key, basePath), "/")
	for depth := 0; depth < len(fieldPath); depth++ {
		cfg = cfg.FieldByName(fieldPath[depth])
		if !cfg.IsValid() {
			log.Println("grc: config field not found:", pair.Key)
			return nil
		}
	}
	// Try DynamicValue.
	if rc.updateDynamicValue(item.Value, cfg) {
		return nil
	}
	// Try StaticValue.
	if forUpdate || rc.setStaticValue(item.Value, cfg, false) {
		return nil
	}
	log.Println("grc: config not updated:", pair.Key, pair.Value)
	return nil
}

func (rc *RemoteConfig) loadUniqueID(node *Node) error {
	ops := Node{}
	opsKey := backend.ServiceOpsKey(rc.path, node.Service, node.Address)
	v, err := rc.provider.Get(opsKey, false)
	if err != nil {
		return err
	}
	if len(v) > 0 {
		err := json.Unmarshal([]byte(v[0].Value), &ops)
		if err != nil {
			return err
		}
		node.UniqueID = ops.UniqueID
		node.Weight = ops.Weight
		return nil
	}

	ops.Weight = node.Weight
	ops.UniqueID, err = rc.provider.Incr(backend.ServiceNodeIDIncrKey(rc.path, node.Service))
	if err != nil {
		return err
	}
	node.UniqueID = ops.UniqueID
	return rc.provider.Set(opsKey, ops.String(), 0)
}

func (rc *RemoteConfig) getServices(basePath string) error {
	services := make(map[string]Nodes)
	kvs, err := rc.provider.Get(basePath, true)
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		var n Node
		err := json.Unmarshal([]byte(kv.Value), &n)
		if err != nil {
			return err
		}
		svc, ok := services[n.Service]
		if !ok {
			svc = make(Nodes)
			services[n.Service] = svc
		}
		svc[n.Address] = &n
	}
	for name, svc := range services {
		rc.svc.Store(name, svc)
	}
	return nil
}

func (rc *RemoteConfig) updateService(basePath, service string) error {
	kvs, err := rc.provider.Get(basePath+service+"/", true)
	if err != nil {
		return err
	}
	svc := make(Nodes, len(kvs))
	for _, kv := range kvs {
		var n Node
		err := json.Unmarshal([]byte(kv.Value), &n)
		if err != nil {
			return err
		}
		svc[n.Address] = &n
	}
	rc.svc.Store(service, svc)
	return nil
}

func (rc *RemoteConfig) watchServiceEvent(basePath string, ch backend.EventChan) {
	var (
		err error
	)

	for {
		select {
		case <-rc.ctx.Done():
			err = rc.provider.Close()
			if err != nil {
				log.Println("grc: stopping.. close provider failed", err)
			}
			return

		case evt := <-ch:
			if evt.Type == backend.Reset {
				err = rc.getServices(basePath)
			} else {
				paths := strings.Split(strings.TrimPrefix(evt.Key, basePath), "/")
				err = rc.updateService(basePath, paths[0])
			}
			if err != nil {
				log.Println("grc: watchServiceEvent failed:", err.Error(), evt.Type, evt.Key)
			}
		}
	}
}
