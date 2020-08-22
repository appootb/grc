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

func WithNodeAddress(addr string) NodeOption {
	return func(node *Node) {
		node.Address = addr
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
	TTL      time.Duration     `json:"ttl"`
	Service  string            `json:"service"`
	Address  string            `json:"address"`
	Weight   int               `json:"weight"`
	Metadata map[string]string `json:"metadata"`
}

func (n Node) String() string {
	v, _ := json.Marshal(n)
	return string(v)
}

type Nodes map[string]*Node
