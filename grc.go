package grc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/appootb/grc/backend"
	"github.com/appootb/grc/backend/etcd"
	"github.com/appootb/grc/backend/memory"
)

var (
	ErrUnknownProvider = errors.New("unknown provider type")
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

type ProviderType string

const (
	Memory ProviderType = backend.Memory
	Etcd   ProviderType = backend.Etcd
)

type RemoteConfig struct {
	svc      sync.Map
	path     string
	ctx      context.Context
	provider backend.Provider
}

func New(ctx context.Context, pt ProviderType, endPoint, user, password, path string) (*RemoteConfig, error) {
	var (
		err      error
		provider backend.Provider
	)

	switch pt {
	case backend.Memory:
		provider, err = memory.NewProvider()
	case backend.Etcd:
		provider, err = etcd.NewProvider(ctx, endPoint, user, password)
	default:
		err = ErrUnknownProvider
	}

	if err != nil {
		return nil, err
	}
	return NewWithProvider(ctx, provider, path)
}

func NewWithProvider(ctx context.Context, provider backend.Provider, path string) (*RemoteConfig, error) {
	rc := &RemoteConfig{
		path:     path,
		ctx:      ctx,
		provider: provider,
	}
	basePath := fmt.Sprintf("%s/%s/", path, ServicePrefix)
	// Watch for service nodes updated.
	evtChan := rc.provider.Watch(basePath, true)
	// Get services.
	if err := rc.getServices(basePath); err != nil {
		return nil, err
	}
	go rc.watchServiceEvent(basePath, evtChan)
	return rc, nil
}

func (rc *RemoteConfig) RegisterNode(service, nodeID string, ttl time.Duration) error {
	if ttl < time.Second {
		ttl = time.Second
	}
	node := &DiscoveryNode{
		NodeID: nodeID,
	}
	key := fmt.Sprintf("%s/%s/%s/", rc.path, ServicePrefix, service)
	return rc.provider.KeepAlive(key, node.String(), ttl)
}

func (rc *RemoteConfig) GetService(service string) map[string]*DiscoveryNode {
	nodes, ok := rc.svc.Load(service)
	if !ok {
		return map[string]*DiscoveryNode{}
	}
	return nodes.(map[string]*DiscoveryNode)
}

func (rc *RemoteConfig) RegisterConfig(service string, v interface{}) error {
	cfg := reflect.ValueOf(v)
	if cfg.Kind() != reflect.Ptr || cfg.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(cfg)}
	}

	basePath := fmt.Sprintf("%s/%s/%s/", rc.path, ConfigPrefix, service)
	// Create default config value if not exist.
	if kvs, err := rc.provider.Get(basePath, true); err != nil {
		return err
	} else if len(kvs) == 0 {
		kvs := parseConfig(reflect.TypeOf(v), "").BackendKVs(basePath)
		for k, v := range kvs {
			err := rc.provider.Set(k, v, 0)
			if err != nil {
				return err
			}
		}
	}

	// Watch for config updated.
	evtChan := rc.provider.Watch(basePath, true)
	// Initialize the config.
	if err := rc.getConfig(basePath, configElem(cfg)); err != nil {
		return err
	}
	go rc.watchConfigEvent(basePath, evtChan, configElem(cfg))
	return nil
}

func (rc *RemoteConfig) getConfig(basePath string, cfg reflect.Value) error {
	kvs, err := rc.provider.Get(basePath, true)
	if err != nil {
		return err
	}
	for _, pair := range kvs {
		err := rc.setConfig(basePath, pair, cfg, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *RemoteConfig) watchConfigEvent(basePath string, ch backend.EventChan, cfg reflect.Value) {
	for {
		select {
		case <-rc.ctx.Done():
			return

		case evt := <-ch:
			err := rc.setConfig(basePath, &evt.KVPair, cfg, true)
			if err != nil {
				log.Println("grc: updateConfig failed:", err.Error())
			}
		}
	}
}

func (rc *RemoteConfig) setConfig(basePath string, pair *backend.KVPair, cfg reflect.Value, forUpdate bool) error {
	var item ConfigItem
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
	// Try set value as base type.
	if setBaseTypeValue(item.Value, cfg) {
		return nil
	}
	// Try system type.
	if forUpdate || setSystemTypeValue(item.Value, cfg, false) {
		return nil
	}
	log.Println("grc: config not updated:", pair.Key, pair.Value)
	return nil
}

func (rc *RemoteConfig) getServices(basePath string) error {
	services := make(map[string]map[string]*DiscoveryNode)
	kvs, err := rc.provider.Get(basePath, true)
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		var n DiscoveryNode
		err := json.Unmarshal([]byte(kv.Value), &n)
		if err != nil {
			return err
		}
		svc, ok := services[n.Service]
		if !ok {
			svc = make(map[string]*DiscoveryNode)
			services[n.Service] = svc
		}
		svc[n.NodeID] = &n
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
	svc := make(map[string]*DiscoveryNode, len(kvs))
	for _, kv := range kvs {
		var n DiscoveryNode
		err := json.Unmarshal([]byte(kv.Value), &n)
		if err != nil {
			return err
		}
		svc[n.NodeID] = &n
	}
	rc.svc.Store(service, svc)
	return nil
}

func (rc *RemoteConfig) watchServiceEvent(basePath string, ch backend.EventChan) {
	for {
		select {
		case <-rc.ctx.Done():
			err := rc.provider.Close()
			if err != nil {
				log.Println("grc: stopping.. close provider failed", err)
			}
			return

		case evt := <-ch:
			paths := strings.Split(strings.TrimPrefix(evt.Key, basePath), "/")
			err := rc.updateService(basePath, paths[0])
			if err != nil {
				log.Println("grc: getService failed:", err.Error())
			}
		}
	}
}
