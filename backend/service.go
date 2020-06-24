package backend

import (
	"encoding/json"
	"fmt"
)

const (
	ServicePrefix       = "service"
	TrafficWeightPrefix = "weight"

	DefaultTrafficWeight = 1
)

func ServiceDiscoveryPrefixKey(path string) string {
	return fmt.Sprintf("%s/%s/", path, ServicePrefix)
}

func ServiceDiscoveryKey(path, service, node string) string {
	return fmt.Sprintf("%s/%s/%s/%s", path, ServicePrefix, service, node)
}

func TrafficWeightKey(path, service, node string) string {
	return fmt.Sprintf("%s/%s/%s/%s", path, TrafficWeightPrefix, service, node)
}

type ServiceNode struct {
	Service  string `json:"service"`
	NodeAddr string `json:"node_addr"`
	Weight   int    `json:"weight"`
}

func (n ServiceNode) String() string {
	v, _ := json.Marshal(n)
	return string(v)
}

type ServiceNodes map[string]int
