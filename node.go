package grc

import (
	"encoding/json"
	"time"
)

type NodeOption func(*Node)

func WithNodeTTL(ttl time.Duration) NodeOption {
	return func(node *Node) {
		node.TTL = ttl
		if node.TTL < time.Second {
			node.TTL = time.Second
		}
	}
}

func WithOpsConfig() NodeOption {
	return func(node *Node) {
		node.ops = true
	}
}

func WithNodeWeight(weight int) NodeOption {
	return func(node *Node) {
		node.Weight = weight
	}
}

func WithNodeMetadata(md map[string]string) NodeOption {
	return func(node *Node) {
		node.Metadata = md
	}
}

type Node struct {
	ops bool

	TTL      time.Duration     `json:"ttl,omitempty"`
	UniqueID int64             `json:"unique_id,omitempty"`
	Service  string            `json:"service,omitempty"`
	Address  string            `json:"address,omitempty"`
	Weight   int               `json:"weight,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (n Node) String() string {
	v, _ := json.Marshal(n)
	return string(v)
}

type Nodes map[string]*Node
