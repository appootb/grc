package grc

import "encoding/json"

const (
	ServicePrefix = "service"
)

type DiscoveryNode struct {
	Service string `json:"service"`
	NodeID  string `json:"node_id"`
	Weight  int    `json:"weight"`
}

func (n DiscoveryNode) String() string {
	v, _ := json.Marshal(n)
	return string(v)
}
