package backend

import (
	"time"
)

const (
	Memory = "memory" // For debug or test usage.
	Etcd   = "etcd"
)

const (
	DialTimeout  = time.Second * 3
	RetryTimeout = time.Second * 3
	ReadTimeout  = time.Second * 3
	WriteTimeout = time.Second * 3
)

type EventType string

const (
	Put    EventType = "put"
	Delete EventType = "delete"
	Reset  EventType = "reset"
)

type KVPair struct {
	Key   string
	Value string
}

type KVPairs []*KVPair

type WatchEvent struct {
	KVPair
	Type EventType
}

type EventChan chan *WatchEvent

const (
	DefaultChanLen = 100
)

// Provider interface.
type Provider interface {
	// Type returns the provider type.
	Type() string

	// Set value for the specified key with a specified ttl.
	Set(key, value string, ttl time.Duration) error

	// Get the value of the specified key or directory.
	Get(key string, dir bool) (KVPairs, error)

	// Incr invokes an atomic value increase for the specified key.
	Incr(key string) (int64, error)

	// Delete the specified key or directory.
	Delete(key string, dir bool) error

	// Watch for changes of the specified key or directory.
	Watch(key string, dir bool) EventChan

	// KeepAlive sets value and updates the ttl for the specified key.
	KeepAlive(key, value string, ttl time.Duration) error

	// Close the provider connection.
	Close() error
}
